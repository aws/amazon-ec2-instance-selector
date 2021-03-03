package ec2pricing

import (
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/pricing"
	"github.com/aws/aws-sdk-go/service/pricing/pricingiface"
)

const (
	defaultSpotDaysBack = 30
	productDescription  = "Linux/UNIX (Amazon VPC)"
	serviceCode         = "AmazonEC2"
)

// EC2Pricing is the public struct to interface with AWS pricing APIs
type EC2Pricing struct {
	PricingClient pricingiface.PricingAPI
	EC2Client     ec2iface.EC2API
	AWSSession    *session.Session
	cache         map[string]float64
	spotCache     map[string]map[string][]spotPricingEntry
}

// EC2PricingIface is the EC2Pricing interface mainly used to mock out ec2pricing during testing
type EC2PricingIface interface {
	GetOndemandInstanceTypeCost(instanceType string) (float64, error)
	GetSpotInstanceTypeNDayAvgCost(instanceType string, availabilityZones []string, days int) (float64, error)
	HydrateOndemandCache() error
	HydrateSpotCache(days int) error
}

type spotPricingEntry struct {
	Timestamp time.Time
	SpotPrice float64
}

// New creates an instance of instance-selector EC2Pricing
func New(sess *session.Session) *EC2Pricing {
	return &EC2Pricing{
		PricingClient: pricing.New(sess),
		EC2Client:     ec2.New(sess),
		AWSSession:    sess,
	}
}

// GetSpotInstanceTypeNDayAvgCost retrieves the spot price history for a given AZ from the past N days and averages the price
// Passing an empty list for availabilityZones will retrieve avg cost for all AZs in the current AWSSession's region
func (p *EC2Pricing) GetSpotInstanceTypeNDayAvgCost(instanceType string, availabilityZones []string, days int) (float64, error) {
	endTime := time.Now().UTC()
	startTime := endTime.Add(time.Hour * time.Duration(24*-1*days))

	spotPriceHistInput := ec2.DescribeSpotPriceHistoryInput{
		ProductDescriptions: []*string{aws.String(productDescription)},
		StartTime:           &startTime,
		EndTime:             &endTime,
		InstanceTypes:       []*string{&instanceType},
	}
	zoneToPriceEntries := map[string][]spotPricingEntry{}

	if _, ok := p.spotCache[instanceType]; !ok {
		var processingErr error
		err := p.EC2Client.DescribeSpotPriceHistoryPages(&spotPriceHistInput, func(dspho *ec2.DescribeSpotPriceHistoryOutput, b bool) bool {
			for _, history := range dspho.SpotPriceHistory {
				var spotPrice float64
				spotPrice, processingErr = strconv.ParseFloat(*history.SpotPrice, 64)
				zone := *history.AvailabilityZone

				zoneToPriceEntries[zone] = append(zoneToPriceEntries[zone], spotPricingEntry{
					Timestamp: *history.Timestamp,
					SpotPrice: spotPrice,
				})
			}
			return true
		})
		if err != nil {
			return float64(0), err
		}
		if processingErr != nil {
			return float64(0), processingErr
		}
	} else {
		for zone, priceEntries := range p.spotCache[instanceType] {
			for _, entry := range priceEntries {
				zoneToPriceEntries[zone] = append(zoneToPriceEntries[zone], spotPricingEntry{
					Timestamp: entry.Timestamp,
					SpotPrice: entry.SpotPrice,
				})
			}
		}
	}

	aggregateZonePriceSum := float64(0)
	numOfZones := 0
	for zone, priceEntries := range zoneToPriceEntries {
		if len(availabilityZones) != 0 {
			if !strings.Contains(strings.Join(availabilityZones, " "), zone) {
				continue
			}
		}
		numOfZones++
		aggregateZonePriceSum += p.calculateSpotAggregate(priceEntries)
	}

	return (aggregateZonePriceSum / float64(numOfZones)), nil
}

func (p *EC2Pricing) calculateSpotAggregate(spotPriceEntries []spotPricingEntry) float64 {
	if len(spotPriceEntries) == 0 {
		return 0.0
	}
	// Sort slice by timestamp in decending order from the end time (most likely, now)
	sort.Slice(spotPriceEntries, func(i, j int) bool {
		return spotPriceEntries[i].Timestamp.After(spotPriceEntries[j].Timestamp)
	})

	endTime := spotPriceEntries[0].Timestamp
	startTime := spotPriceEntries[len(spotPriceEntries)-1].Timestamp
	totalDuration := endTime.Sub(startTime).Minutes()

	priceSum := float64(0)
	for i, entry := range spotPriceEntries {
		duration := spotPriceEntries[int(math.Max(float64(i-1), 0))].Timestamp.Sub(entry.Timestamp).Minutes()
		priceSum += duration * entry.SpotPrice
	}
	return (priceSum / totalDuration)
}

// GetOndemandInstanceTypeCost retrieves the on-demand hourly cost for the specified instance type
func (p *EC2Pricing) GetOndemandInstanceTypeCost(instanceType string) (float64, error) {
	regionDescription := p.getRegionForPricingAPI()
	// TODO: mac.metal instances cannot be found with the below filters
	productInput := pricing.GetProductsInput{
		ServiceCode: aws.String(serviceCode),
		Filters: []*pricing.Filter{
			{Type: aws.String(pricing.FilterTypeTermMatch), Field: aws.String("ServiceCode"), Value: aws.String(serviceCode)},
			{Type: aws.String(pricing.FilterTypeTermMatch), Field: aws.String("operatingSystem"), Value: aws.String("linux")},
			{Type: aws.String(pricing.FilterTypeTermMatch), Field: aws.String("location"), Value: aws.String(regionDescription)},
			{Type: aws.String(pricing.FilterTypeTermMatch), Field: aws.String("capacitystatus"), Value: aws.String("used")},
			{Type: aws.String(pricing.FilterTypeTermMatch), Field: aws.String("preInstalledSw"), Value: aws.String("NA")},
			{Type: aws.String(pricing.FilterTypeTermMatch), Field: aws.String("tenancy"), Value: aws.String("shared")},
			{Type: aws.String(pricing.FilterTypeTermMatch), Field: aws.String("instanceType"), Value: aws.String(instanceType)},
		},
	}

	// Check cache first and return it if available
	if price, ok := p.cache[instanceType]; ok {
		return price, nil
	}

	pricePerUnitInUSD := float64(-1)
	err := p.PricingClient.GetProductsPages(&productInput, func(pricingOutput *pricing.GetProductsOutput, nextPage bool) bool {
		var err error
		for _, priceDoc := range pricingOutput.PriceList {
			_, pricePerUnitInUSD, err = parseOndemandUnitPrice(priceDoc)
		}
		if err != nil {
			// keep going through pages if we can't parse the pricing doc
			return true
		}
		return false
	})
	if err != nil {
		return -1, err
	}
	return pricePerUnitInUSD, nil
}

// HydrateSpotCache makes a bulk request to the spot-pricing-history api to retrieve all instance type pricing and stores them in a local cache
// If HydrateSpotCache is called more than once, the cache will be fully refreshed
// There is no TTL on cache entries
// You'll only want to use this if you don't mind a long startup time (around 30 seconds) and will query the cache often after that.
func (p *EC2Pricing) HydrateSpotCache(days int) error {
	newCache := map[string]map[string][]spotPricingEntry{}

	endTime := time.Now().UTC()
	startTime := endTime.Add(time.Hour * time.Duration(24*-1*days))
	spotPriceHistInput := ec2.DescribeSpotPriceHistoryInput{
		ProductDescriptions: []*string{aws.String(productDescription)},
		StartTime:           &startTime,
		EndTime:             &endTime,
	}
	var processingErr error
	err := p.EC2Client.DescribeSpotPriceHistoryPages(&spotPriceHistInput, func(dspho *ec2.DescribeSpotPriceHistoryOutput, b bool) bool {
		for _, history := range dspho.SpotPriceHistory {
			var spotPrice float64
			spotPrice, processingErr = strconv.ParseFloat(*history.SpotPrice, 64)
			instanceType := *history.InstanceType
			zone := *history.AvailabilityZone
			if _, ok := newCache[instanceType]; !ok {
				newCache[instanceType] = map[string][]spotPricingEntry{}
			}
			newCache[instanceType][zone] = append(newCache[instanceType][zone], spotPricingEntry{
				Timestamp: *history.Timestamp,
				SpotPrice: spotPrice,
			})
		}
		return true
	})
	if err != nil {
		return err
	}
	p.spotCache = newCache
	return processingErr
}

// HydrateOndemandCache makes a bulk request to the pricing api to retrieve all instance type pricing and stores them in a local cache
// If HydrateOndemandCache is called more than once, the cache will be fully refreshed
// There is no TTL on cache entries
func (p *EC2Pricing) HydrateOndemandCache() error {
	if p.cache == nil {
		p.cache = make(map[string]float64)
	}
	regionDescription := p.getRegionForPricingAPI()
	productInput := pricing.GetProductsInput{
		ServiceCode: aws.String(serviceCode),
		Filters: []*pricing.Filter{
			{Type: aws.String(pricing.FilterTypeTermMatch), Field: aws.String("ServiceCode"), Value: aws.String(serviceCode)},
			{Type: aws.String(pricing.FilterTypeTermMatch), Field: aws.String("operatingSystem"), Value: aws.String("linux")},
			{Type: aws.String(pricing.FilterTypeTermMatch), Field: aws.String("location"), Value: aws.String(regionDescription)},
			{Type: aws.String(pricing.FilterTypeTermMatch), Field: aws.String("capacitystatus"), Value: aws.String("used")},
			{Type: aws.String(pricing.FilterTypeTermMatch), Field: aws.String("preInstalledSw"), Value: aws.String("NA")},
			{Type: aws.String(pricing.FilterTypeTermMatch), Field: aws.String("tenancy"), Value: aws.String("shared")},
		},
	}
	err := p.PricingClient.GetProductsPages(&productInput, func(pricingOutput *pricing.GetProductsOutput, nextPage bool) bool {
		for _, priceDoc := range pricingOutput.PriceList {
			instanceTypeName, price, err := parseOndemandUnitPrice(priceDoc)
			if err != nil {
				continue
			}
			p.cache[instanceTypeName] = price
		}
		return true
	})
	return err
}

// getRegionForPricingAPI attempts to retrieve the region description based on the AWS session used to create
// the ec2pricing struct. It then uses the endpoints package in the aws sdk to retrieve the region description
// This is necessary because the pricing API uses the region description rather than a region ID
func (p *EC2Pricing) getRegionForPricingAPI() string {
	endpointResolver := endpoints.DefaultResolver()
	partitions := endpointResolver.(endpoints.EnumPartitions).Partitions()

	// use us-east-1 as the default
	regionDescription := "US East (N. Virginia)"
	for _, partition := range partitions {
		regions := partition.Regions()
		if region, ok := regions[*p.AWSSession.Config.Region]; ok {
			regionDescription = region.Description()
		}
	}
	return regionDescription
}

// parseOndemandUnitPrice takes a priceList from the pricing API and parses its weirdness
func parseOndemandUnitPrice(priceList aws.JSONValue) (string, float64, error) {
	// TODO: this could probably be cleaned up a bit by adding a couple structs with json tags
	//       We still need to some weird for-loops to get at elements under json keys that are IDs...
	//       But it would probably be cleaner than this.
	attributes, ok := priceList["product"].(map[string]interface{})["attributes"]
	if !ok {
		return "", float64(-1.0), fmt.Errorf("Unable to find product attributes")
	}
	instanceTypeName, ok := attributes.(map[string]interface{})["instanceType"].(string)
	if !ok {
		return "", float64(-1.0), fmt.Errorf("Unable to find instance type name from product attributes")
	}
	terms, ok := priceList["terms"]
	if !ok {
		return instanceTypeName, float64(-1.0), fmt.Errorf("Unable to find pricing terms")
	}
	ondemandTerms, ok := terms.(map[string]interface{})["OnDemand"]
	if !ok {
		return instanceTypeName, float64(-1.0), fmt.Errorf("Unable to find on-demand pricing terms")
	}
	for _, priceDimensions := range ondemandTerms.(map[string]interface{}) {
		dim, ok := priceDimensions.(map[string]interface{})["priceDimensions"]
		if !ok {
			return instanceTypeName, float64(-1.0), fmt.Errorf("Unable to find on-demand pricing dimensions")
		}
		for _, dimension := range dim.(map[string]interface{}) {
			dims := dimension.(map[string]interface{})
			pricePerUnit, ok := dims["pricePerUnit"]
			if !ok {
				return instanceTypeName, float64(-1.0), fmt.Errorf("Unable to find on-demand price per unit in pricing dimensions")
			}
			pricePerUnitInUSDStr, ok := pricePerUnit.(map[string]interface{})["USD"]
			if !ok {
				return instanceTypeName, float64(-1.0), fmt.Errorf("Unable to find on-demand price per unit in USD")
			}
			var err error
			pricePerUnitInUSD, err := strconv.ParseFloat(pricePerUnitInUSDStr.(string), 64)
			if err != nil {
				return instanceTypeName, float64(-1.0), fmt.Errorf("Could not convert price per unit in USD to a float64")
			}
			return instanceTypeName, pricePerUnitInUSD, nil
		}
	}
	return instanceTypeName, float64(-1.0), fmt.Errorf("Unable to parse pricing doc")
}
