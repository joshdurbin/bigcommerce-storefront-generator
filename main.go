package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/brianvoe/gofakeit/v7"
)

const (
	// Configuration constants
	StoreHash       = "yourstorehash"
	AuthToken       = "yourauthtoken"
	NumCategories   = 10
	NumBrands       = 5
	NumProducts     = 30
	NumCustomFields = 2
	MaxVariants     = 3
	MaxOptions      = 2
	MaxImages       = 3
	MaxVideos       = 1
	MaxReviews      = 5
)

func main() {
	// Seed the random generator
	gofakeit.Seed(time.Now().UnixNano())
	rand.Seed(time.Now().UnixNano())

	// Initialize the BigCommerce client
	client := NewClient(StoreHash, AuthToken)

	// Create context
	ctx := context.Background()

	// Generate and create categories
	categories := generateCategories()
	categoryIDs, err := createCategories(ctx, client, categories)
	if err != nil {
		log.Fatalf("Failed to create categories: %v", err)
	}
	log.Printf("Created %d categories", len(categoryIDs))

	// Generate and create brands
	brands := generateBrands()
	brandIDs, err := createBrands(ctx, client, brands)
	if err != nil {
		log.Fatalf("Failed to create brands: %v", err)
	}
	log.Printf("Created %d brands", len(brandIDs))

	// Generate and create products
	products := generateProducts(categoryIDs, brandIDs)
	productIDs, err := createProducts(ctx, client, products)
	if err != nil {
		log.Fatalf("Failed to create products: %v", err)
	}
	log.Printf("Created %d products", len(productIDs))

	// For each product, add additional data
	for _, productID := range productIDs {
		// Add custom fields
		if err := addCustomFields(ctx, client, productID); err != nil {
			log.Printf("Failed to add custom fields for product %d: %v", productID, err)
			continue
		}

		// Add images
		if err := addProductImages(ctx, client, productID); err != nil {
			log.Printf("Failed to add images for product %d: %v", productID, err)
		}

		// Add videos
		if err := addProductVideos(ctx, client, productID); err != nil {
			log.Printf("Failed to add videos for product %d: %v", productID, err)
		}

		// Add options and variants
		if err := addOptionsAndVariants(ctx, client, productID); err != nil {
			log.Printf("Failed to add options and variants for product %d: %v", productID, err)
		}

		// Add reviews
		if err := addProductReviews(ctx, client, productID); err != nil {
			log.Printf("Failed to add reviews for product %d: %v", productID, err)
		}

		// Add bulk pricing rules
		if err := addBulkPricingRules(ctx, client, productID); err != nil {
			log.Printf("Failed to add bulk pricing rules for product %d: %v", productID, err)
		}
	}

	log.Println("Finished creating store catalog data!")
}

func generateCategories() []Category {
	categories := make([]Category, NumCategories)

	// First category is top-level
	categories[0] = Category{
		ParentID:        0,
		Name:            gofakeit.ProductCategory(),
		Description:     gofakeit.ProductDescription(),
		SortOrder:       0,
		PageTitle:       gofakeit.Sentence(3),
		MetaKeywords:    []string{gofakeit.Word(), gofakeit.Word(), gofakeit.Word()},
		MetaDescription: gofakeit.Paragraph(1, 2, 3, " "),
		LayoutFile:      "category.html",
		IsVisible:       true,
		ImageURL:        "https://images.pexels.com/photos/45201/kitty-cat-kitten-pet-45201.jpeg",
	}

	// Rest can be children
	for i := 1; i < NumCategories; i++ {
		parentID := 0
		if i > 2 { // Some categories have parents
			parentID = rand.Intn(i) // Parent must have lower ID
		}

		categories[i] = Category{
			ParentID:        parentID,
			Name:            gofakeit.ProductCategory(),
			Description:     gofakeit.ProductDescription(),
			SortOrder:       i,
			PageTitle:       gofakeit.Sentence(3),
			MetaKeywords:    []string{gofakeit.Word(), gofakeit.Word(), gofakeit.Word()},
			MetaDescription: gofakeit.Paragraph(1, 2, 3, " "),
			LayoutFile:      "category.html",
			IsVisible:       rand.Float32() > 0.1, // 90% visible
			ImageURL:        "https://images.pexels.com/photos/45201/kitty-cat-kitten-pet-45201.jpeg",
		}
	}

	return categories
}

func createCategories(ctx context.Context, client *Client, categories []Category) ([]int, error) {
	categoryIDs := make([]int, 0, len(categories))

	for _, category := range categories {
		response, err := client.Categories.CreateContext(ctx, &category)
		if err != nil {
			return categoryIDs, fmt.Errorf("failed to create category: %v", err)
		}
		categoryIDs = append(categoryIDs, response.Data.ID)
		log.Printf("Created category: %s (ID: %d)", category.Name, response.Data.ID)
	}

	return categoryIDs, nil
}

func generateBrands() []Brand {
	brands := make([]Brand, NumBrands)

	for i := 0; i < NumBrands; i++ {
		brandName := gofakeit.Company()
		brands[i] = Brand{
			Name:            brandName,
			PageTitle:       brandName + " Products",
			MetaKeywords:    []string{brandName, gofakeit.Word(), gofakeit.Word()},
			MetaDescription: gofakeit.Paragraph(1, 2, 3, " "),
			ImageURL:        "https://images.pexels.com/photos/45201/kitty-cat-kitten-pet-45201.jpeg",
			SearchKeywords:  gofakeit.Word() + ", " + gofakeit.Word(),
		}
	}

	return brands
}

func createBrands(ctx context.Context, client *Client, brands []Brand) ([]int, error) {
	brandIDs := make([]int, 0, len(brands))

	for _, brand := range brands {
		response, err := client.Brands.CreateContext(ctx, &brand)
		if err != nil {
			return brandIDs, fmt.Errorf("failed to create brand: %v", err)
		}
		brandIDs = append(brandIDs, response.Data.ID)
		log.Printf("Created brand: %s (ID: %d)", brand.Name, response.Data.ID)
	}

	return brandIDs, nil
}

func generateProducts(categoryIDs, brandIDs []int) []Product {
	products := make([]Product, NumProducts)

	for i := 0; i < NumProducts; i++ {
		// Select random categories (1-3)
		numCats := rand.Intn(3) + 1
		categories := make([]int, 0, numCats)
		for j := 0; j < numCats; j++ {
			catID := categoryIDs[rand.Intn(len(categoryIDs))]
			// Check if already added
			alreadyAdded := false
			for _, c := range categories {
				if c == catID {
					alreadyAdded = true
					break
				}
			}
			if !alreadyAdded {
				categories = append(categories, catID)
			}
		}

		// Select random brand
		brandID := brandIDs[rand.Intn(len(brandIDs))]

		// Generate product details
		name := gofakeit.ProductName()
		price := gofakeit.Price(10, 1000)
		weight := gofakeit.Float64Range(0.1, 25)
		inventory := rand.Intn(100)

		products[i] = Product{
			Name:              name,
			Type:              "physical",
			SKU:               gofakeit.UUID(),
			Description:       gofakeit.ProductDescription(),
			Weight:            weight,
			Width:             gofakeit.Float64Range(1, 50),
			Depth:             gofakeit.Float64Range(1, 50),
			Height:            gofakeit.Float64Range(1, 50),
			Price:             price,
			CostPrice:         price * 0.6, // 60% of retail
			RetailPrice:       price * 1.2, // 20% markup
			SalePrice:         price * 0.9, // 10% discount
			Categories:        categories,
			BrandID:           brandID,
			InventoryLevel:    inventory,
			InventoryWarning:  10,
			InventoryTracking: "product",
			IsVisible:         true,
			IsFeatured:        rand.Float32() < 0.2, // 20% featured
			Warranty:          gofakeit.Sentence(10),
			BinPickingNumber:  gofakeit.DigitN(6),
			UPC:               gofakeit.DigitN(12),
			MPN:               fmt.Sprintf("MPN-%s", gofakeit.DigitN(8)),
			GTIN:              gofakeit.DigitN(14),
			SearchKeywords:    gofakeit.Word() + ", " + gofakeit.Word() + ", " + gofakeit.Word(),
			Availability:      "available",
			AvailabilityDesc:  "Usually ships in 1-2 business days",
			SortOrder:         i,
			Condition:         "New",
			IsConditionShown:  true,
			OrderQuantityMin:  1,
			OrderQuantityMax:  10,
			PageTitle:         name,
			MetaKeywords:      []string{gofakeit.Word(), gofakeit.Word(), gofakeit.Word()},
			MetaDescription:   gofakeit.Paragraph(1, 2, 3, " "),
			//CustomURL:         &CustomURL{URL: gofakeit.URL()},
			OpenGraphType:  "product",
			OpenGraphTitle: name,
			OpenGraphDesc:  gofakeit.Sentence(5),
		}
	}

	return products
}

func createProducts(ctx context.Context, client *Client, products []Product) ([]int, error) {
	productIDs := make([]int, 0, len(products))

	for _, product := range products {
		response, err := client.Products.CreateContext(ctx, &product)
		if err != nil {
			return productIDs, fmt.Errorf("failed to create product: %v", err)
		}
		productIDs = append(productIDs, response.Data.ID)
		log.Printf("Created product: %s (ID: %d)", product.Name, response.Data.ID)
	}

	return productIDs, nil
}

func addCustomFields(ctx context.Context, client *Client, productID int) error {
	for i := 0; i < NumCustomFields; i++ {
		field := &CustomField{
			Name:  gofakeit.Word() + " Info",
			Value: gofakeit.Sentence(5),
		}

		_, err := client.CustomFields.CreateContext(ctx, productID, field)
		if err != nil {
			return fmt.Errorf("failed to create custom field: %v", err)
		}
	}

	return nil
}

func addProductImages(ctx context.Context, client *Client, productID int) error {
	numImages := rand.Intn(MaxImages) + 1
	for i := 0; i < numImages; i++ {
		image := &ProductImage{
			ImageFile:   "https://images.pexels.com/photos/45201/kitty-cat-kitten-pet-45201.jpeg",
			IsThumbnail: i == 0,
			SortOrder:   i,
			Description: gofakeit.Sentence(5),
		}

		_, err := client.ProductImages.CreateContext(ctx, productID, image)
		if err != nil {
			return fmt.Errorf("failed to create product image: %v", err)
		}
	}

	return nil
}

func addProductVideos(ctx context.Context, client *Client, productID int) error {
	numVideos := rand.Intn(MaxVideos + 1)

	if numVideos == 0 {
		return nil
	}

	for i := 0; i < numVideos; i++ {
		videoID := gofakeit.UUID() // Using UUID as dummy YouTube video ID

		video := &ProductVideo{
			Title:       gofakeit.ProductName() + " Video",
			Description: gofakeit.Sentence(10),
			SortOrder:   i,
			Type:        "youtube",
			VideoID:     videoID,
		}

		_, err := client.ProductVideos.CreateContext(ctx, productID, video)
		if err != nil {
			return fmt.Errorf("failed to create product video: %v", err)
		}
	}

	return nil
}

func addOptionsAndVariants(ctx context.Context, client *Client, productID int) error {
	numOptions := rand.Intn(MaxOptions + 1)

	if numOptions == 0 {
		return nil
	}

	optionTypes := []string{"dropdown", "radio", "checkbox", "swatch"}
	optionNames := []string{"Color", "Size", "Material", "Style"}

	optionIDs := make([]int, 0, numOptions)
	optionValueMap := make(map[int][]OptionValue)

	// Create options
	for i := 0; i < numOptions; i++ {
		optionType := optionTypes[rand.Intn(len(optionTypes))]
		optionName := optionNames[i%len(optionNames)]

		option := &ProductOption{
			DisplayName: optionName,
			Type:        optionType,
		}

		optionResp, err := client.Options.CreateContext(ctx, productID, option)
		if err != nil {
			return fmt.Errorf("failed to create product option: %v", err)
		}

		optionID := optionResp.Data.ID
		optionIDs = append(optionIDs, optionID)

		// Create option values
		numValues := rand.Intn(3) + 2 // 2-4 values
		values := make([]OptionValue, 0, numValues)

		for j := 0; j < numValues; j++ {
			var value string

			switch optionName {
			case "Color":
				value = gofakeit.Color()
			case "Size":
				sizes := []string{"Small", "Medium", "Large", "X-Large", "XX-Large"}
				value = sizes[j%len(sizes)]
			case "Material":
				materials := []string{"Cotton", "Polyester", "Wool", "Leather", "Silk"}
				value = materials[j%len(materials)]
			default:
				value = gofakeit.Word()
			}

			optionValue := OptionValue{
				OptionID:  optionID,
				Label:     value,
				SortOrder: j,
				Value:     value,
				IsDefault: j == 0,
			}

			valueResp, err := client.Options.CreateOptionValueContext(ctx, productID, optionID, &optionValue)
			if err != nil {
				return fmt.Errorf("failed to create option value: %v", err)
			}

			optionValue.ID = valueResp.Data.ID
			values = append(values, optionValue)
		}

		optionValueMap[optionID] = values
	}

	// Create variants if there are options
	if len(optionIDs) > 0 {
		numVariants := rand.Intn(MaxVariants) + 1

		for i := 0; i < numVariants; i++ {
			// Create variant options
			variantOptions := make([]OptionValue, 0, len(optionIDs))

			for _, optionID := range optionIDs {
				values := optionValueMap[optionID]
				valueIndex := rand.Intn(len(values))
				variantOptions = append(variantOptions, values[valueIndex])
			}

			// Create variant
			variant := &Variant{
				SKU:                   gofakeit.UUID(),
				Price:                 gofakeit.Price(10, 1000),
				Weight:                gofakeit.Float64Range(0.1, 25),
				Depth:                 gofakeit.Float64Range(1, 50),
				Height:                gofakeit.Float64Range(1, 50),
				Width:                 gofakeit.Float64Range(1, 50),
				InventoryLevel:        rand.Intn(100),
				InventoryWarningLevel: 10,
				OptionValues:          variantOptions,
			}

			_, err := client.Variants.CreateContext(ctx, productID, variant)
			if err != nil {
				return fmt.Errorf("failed to create variant: %v", err)
			}
		}
	}

	return nil
}

func addProductReviews(ctx context.Context, client *Client, productID int) error {
	numReviews := rand.Intn(MaxReviews + 1)

	if numReviews == 0 {
		return nil
	}

	for i := 0; i < numReviews; i++ {
		rating := rand.Intn(4) + 2 // Ratings 2-5

		review := &Review{
			Title:  gofakeit.Sentence(3),
			Text:   gofakeit.Paragraph(1, 3, 5, " "),
			Status: "approved",
			Rating: rating,
			Name:   gofakeit.Name(),
			Email:  gofakeit.Email(),
		}

		_, err := client.Reviews.CreateContext(ctx, productID, review)
		if err != nil {
			return fmt.Errorf("failed to create review: %v", err)
		}
	}

	return nil
}

func addBulkPricingRules(ctx context.Context, client *Client, productID int) error {
	// Only add bulk pricing rules to 30% of products
	if rand.Float32() > 0.3 {
		return nil
	}

	// Define some tiers
	tiers := []struct {
		Min    int
		Max    int
		Amount float64
	}{
		{2, 9, 5},
		{10, 19, 10},
		{20, 0, 15}, // 0 max means unlimited
	}

	for _, tier := range tiers {
		rule := &PricingRule{
			QuantityMin: tier.Min,
			QuantityMax: tier.Max,
			Type:        "price", // or "percent"
			Amount:      tier.Amount,
		}

		_, err := client.BulkPricingRules.CreateContext(ctx, productID, rule)
		if err != nil {
			return fmt.Errorf("failed to create bulk pricing rule: %v", err)
		}
	}

	return nil
}
