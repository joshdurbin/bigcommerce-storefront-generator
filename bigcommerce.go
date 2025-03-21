package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	defaultBaseURL = "https://api.bigcommerce.com/stores/"
	apiVersion     = "v3"
	userAgent      = "bigcommerce-go-sdk/1.0"
)

type Client struct {
	client *http.Client

	baseURL *url.URL

	storeHash string
	authToken string

	userAgent string

	Products                  *ProductsService
	Categories                *CategoriesService
	Brands                    *BrandsService
	Variants                  *VariantsService
	ProductImages             *ProductImagesService
	ProductVideos             *VideosService
	Options                   *OptionsService
	Modifiers                 *ModifiersService
	Reviews                   *ReviewsService
	ComplexRules              *ComplexRulesService
	CustomFields              *CustomFieldsService
	Metafields                *MetafieldsService
	Channels                  *ChannelsService
	Summary                   *SummaryService
	RelatedProducts           *RelatedProductsService
	ProductChannelAssignments *ProductChannelAssignmentsService
	ProductCategories         *ProductCategoriesService
	Batch                     *BatchService
	Pricing                   *PricingService
	Inventory                 *InventoryService
	BulkPricingRules          *BulkPricingRulesService
}

func NewClient(storeHash, authToken string) *Client {
	httpClient := &http.Client{
		Timeout: time.Second * 30,
	}

	baseURL, _ := url.Parse(defaultBaseURL + storeHash + "/" + apiVersion + "/")

	c := &Client{
		client:    httpClient,
		baseURL:   baseURL,
		storeHash: storeHash,
		authToken: authToken,
		userAgent: userAgent,
	}

	c.Products = &ProductsService{client: c}
	c.Categories = &CategoriesService{client: c}
	c.Brands = &BrandsService{client: c}
	c.Variants = &VariantsService{client: c}
	c.ProductImages = &ProductImagesService{client: c}
	c.ProductVideos = &VideosService{client: c}
	c.Options = &OptionsService{client: c}
	c.Modifiers = &ModifiersService{client: c}
	c.Reviews = &ReviewsService{client: c}
	c.ComplexRules = &ComplexRulesService{client: c}
	c.CustomFields = &CustomFieldsService{client: c}
	c.Metafields = &MetafieldsService{client: c}
	c.Channels = &ChannelsService{client: c}
	c.Summary = &SummaryService{client: c}
	c.RelatedProducts = &RelatedProductsService{client: c}
	c.ProductChannelAssignments = &ProductChannelAssignmentsService{client: c}
	c.ProductCategories = &ProductCategoriesService{client: c}
	c.Batch = &BatchService{client: c}
	c.Pricing = &PricingService{client: c}
	c.Inventory = &InventoryService{client: c}
	c.BulkPricingRules = &BulkPricingRulesService{client: c}

	return c
}

func (c *Client) NewRequest(ctx context.Context, method, urlStr string, body interface{}) (*http.Request, error) {
	rel, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	u := c.baseURL.ResolveReference(rel)

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Auth-Token", c.authToken)
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return req, nil
}

func (c *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	err = CheckResponse(resp)
	if err != nil {
		return resp, err
	}

	if v != nil {
		if w, ok := v.(io.Writer); ok {
			_, err = io.Copy(w, resp.Body)
		} else {
			err = json.NewDecoder(resp.Body).Decode(v)
		}
	}

	return resp, err
}

type ErrorResponse struct {
	Response *http.Response
	Status   int      `json:"status"`
	Title    string   `json:"title"`
	Type     string   `json:"type"`
	Errors   []string `json:"errors"`
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("%v %v: %d %v - %v",
		e.Response.Request.Method, e.Response.Request.URL,
		e.Response.StatusCode, e.Title, e.Errors)
}

func CheckResponse(r *http.Response) error {
	if c := r.StatusCode; c >= 200 && c <= 299 {
		return nil
	}

	errorResponse := &ErrorResponse{Response: r}
	data, err := io.ReadAll(r.Body)
	if err == nil && len(data) > 0 {
		err := json.Unmarshal(data, errorResponse)
		if err != nil {
			return err
		}
	}

	return errorResponse
}

type QueryParams struct {
	Page         int
	Limit        int
	Direction    string
	Sort         string
	Include      []string
	ID           []int
	IDIn         []int
	IDNotIn      []int
	IDMin        int
	IDMax        int
	IDGreater    int
	IDLess       int
	Name         string
	SKU          string
	Price        float64
	PriceMin     float64
	PriceMax     float64
	Weight       float64
	WeightMin    float64
	WeightMax    float64
	Condition    string
	IsVisible    *bool
	IsFeatured   *bool
	CategoryID   []int
	BrandID      []int
	Keywords     string
	IsActive     *bool
	DateCreated  string
	DateModified string
}

func (q *QueryParams) ToValues() url.Values {
	values := url.Values{}

	if q.Page > 0 {
		values.Add("page", strconv.Itoa(q.Page))
	}

	if q.Limit > 0 {
		values.Add("limit", strconv.Itoa(q.Limit))
	}

	if q.Direction != "" {
		values.Add("direction", q.Direction)
	}

	if q.Sort != "" {
		values.Add("sort", q.Sort)
	}

	for _, include := range q.Include {
		values.Add("include", include)
	}

	for _, id := range q.ID {
		values.Add("id", strconv.Itoa(id))
	}

	for _, id := range q.IDIn {
		values.Add("id:in", strconv.Itoa(id))
	}

	for _, id := range q.IDNotIn {
		values.Add("id:not_in", strconv.Itoa(id))
	}

	if q.IDMin > 0 {
		values.Add("id:min", strconv.Itoa(q.IDMin))
	}

	if q.IDMax > 0 {
		values.Add("id:max", strconv.Itoa(q.IDMax))
	}

	if q.IDGreater > 0 {
		values.Add("id:greater", strconv.Itoa(q.IDGreater))
	}

	if q.IDLess > 0 {
		values.Add("id:less", strconv.Itoa(q.IDLess))
	}

	if q.Name != "" {
		values.Add("name", q.Name)
	}

	if q.SKU != "" {
		values.Add("sku", q.SKU)
	}

	if q.Price > 0 {
		values.Add("price", strconv.FormatFloat(q.Price, 'f', 2, 64))
	}

	if q.PriceMin > 0 {
		values.Add("price:min", strconv.FormatFloat(q.PriceMin, 'f', 2, 64))
	}

	if q.PriceMax > 0 {
		values.Add("price:max", strconv.FormatFloat(q.PriceMax, 'f', 2, 64))
	}

	if q.Weight > 0 {
		values.Add("weight", strconv.FormatFloat(q.Weight, 'f', 2, 64))
	}

	if q.WeightMin > 0 {
		values.Add("weight:min", strconv.FormatFloat(q.WeightMin, 'f', 2, 64))
	}

	if q.WeightMax > 0 {
		values.Add("weight:max", strconv.FormatFloat(q.WeightMax, 'f', 2, 64))
	}

	if q.Condition != "" {
		values.Add("condition", q.Condition)
	}

	if q.IsVisible != nil {
		values.Add("is_visible", strconv.FormatBool(*q.IsVisible))
	}

	if q.IsFeatured != nil {
		values.Add("is_featured", strconv.FormatBool(*q.IsFeatured))
	}

	for _, categoryID := range q.CategoryID {
		values.Add("categories", strconv.Itoa(categoryID))
	}

	for _, brandID := range q.BrandID {
		values.Add("brand_id", strconv.Itoa(brandID))
	}

	if q.Keywords != "" {
		values.Add("keyword", q.Keywords)
	}

	if q.IsActive != nil {
		values.Add("is_active", strconv.FormatBool(*q.IsActive))
	}

	if q.DateCreated != "" {
		values.Add("date_created", q.DateCreated)
	}

	if q.DateModified != "" {
		values.Add("date_modified", q.DateModified)
	}

	return values
}

type Meta struct {
	Pagination struct {
		Total       int `json:"total"`
		Count       int `json:"count"`
		PerPage     int `json:"per_page"`
		CurrentPage int `json:"current_page"`
		TotalPages  int `json:"total_pages"`
		Links       struct {
			Current  string `json:"current"`
			Previous string `json:"previous"`
			Next     string `json:"next"`
		} `json:"links"`
	} `json:"pagination"`
}

type Product struct {
	ID                  int             `json:"id,omitempty"`
	Name                string          `json:"name"`
	Type                string          `json:"type"`
	SKU                 string          `json:"sku,omitempty"`
	Description         string          `json:"description,omitempty"`
	Weight              float64         `json:"weight,omitempty"`
	Width               float64         `json:"width,omitempty"`
	Depth               float64         `json:"depth,omitempty"`
	Height              float64         `json:"height,omitempty"`
	Price               float64         `json:"price"`
	CostPrice           float64         `json:"cost_price,omitempty"`
	RetailPrice         float64         `json:"retail_price,omitempty"`
	SalePrice           float64         `json:"sale_price,omitempty"`
	MapPrice            float64         `json:"map_price,omitempty"`
	TaxClassID          int             `json:"tax_class_id,omitempty"`
	ProductTaxCode      string          `json:"product_tax_code,omitempty"`
	Categories          []int           `json:"categories,omitempty"`
	BrandID             int             `json:"brand_id,omitempty"`
	InventoryLevel      int             `json:"inventory_level,omitempty"`
	InventoryWarning    int             `json:"inventory_warning_level,omitempty"`
	InventoryTracking   string          `json:"inventory_tracking,omitempty"`
	FixedCostShipping   float64         `json:"fixed_cost_shipping_price,omitempty"`
	IsFreeShipping      bool            `json:"is_free_shipping,omitempty"`
	IsVisible           bool            `json:"is_visible,omitempty"`
	IsFeatured          bool            `json:"is_featured,omitempty"`
	RelatedProducts     []int           `json:"related_products,omitempty"`
	Warranty            string          `json:"warranty,omitempty"`
	BinPickingNumber    string          `json:"bin_picking_number,omitempty"`
	Layout              string          `json:"layout_file,omitempty"`
	UPC                 string          `json:"upc,omitempty"`
	MPN                 string          `json:"mpn,omitempty"`
	GTIN                string          `json:"gtin,omitempty"`
	SearchKeywords      string          `json:"search_keywords,omitempty"`
	Availability        string          `json:"availability,omitempty"`
	AvailabilityDesc    string          `json:"availability_description,omitempty"`
	GiftWrappingOpts    string          `json:"gift_wrapping_options_type,omitempty"`
	SortOrder           int             `json:"sort_order,omitempty"`
	Condition           string          `json:"condition,omitempty"`
	IsConditionShown    bool            `json:"is_condition_shown,omitempty"`
	OrderQuantityMin    int             `json:"order_quantity_minimum,omitempty"`
	OrderQuantityMax    int             `json:"order_quantity_maximum,omitempty"`
	PageTitle           string          `json:"page_title,omitempty"`
	MetaKeywords        []string        `json:"meta_keywords,omitempty"`
	MetaDescription     string          `json:"meta_description,omitempty"`
	DateCreated         string          `json:"date_created,omitempty"`
	DateModified        string          `json:"date_modified,omitempty"`
	ViewCount           int             `json:"view_count,omitempty"`
	PreorderReleaseDate string          `json:"preorder_release_date,omitempty"`
	PreorderMessage     string          `json:"preorder_message,omitempty"`
	IsPreorderOnly      bool            `json:"is_preorder_only,omitempty"`
	IsPriceHidden       bool            `json:"is_price_hidden,omitempty"`
	PriceHiddenLabel    string          `json:"price_hidden_label,omitempty"`
	CustomURL           *CustomURL      `json:"custom_url,omitempty"`
	BaseVariantID       int             `json:"base_variant_id,omitempty"`
	OpenGraphType       string          `json:"open_graph_type,omitempty"`
	OpenGraphTitle      string          `json:"open_graph_title,omitempty"`
	OpenGraphDesc       string          `json:"open_graph_description,omitempty"`
	Images              []ProductImage  `json:"images,omitempty"`
	Videos              []ProductVideo  `json:"videos,omitempty"`
	CustomFields        []CustomField   `json:"custom_fields,omitempty"`
	BulkPricingRules    []PricingRule   `json:"bulk_pricing_rules,omitempty"`
	Variants            []Variant       `json:"variants,omitempty"`
	Options             []ProductOption `json:"options,omitempty"`
	Modifiers           []Modifier      `json:"modifiers,omitempty"`
	Reviews             []Review        `json:"reviews,omitempty"`
	ComplexRules        []ComplexRule   `json:"complex_rules,omitempty"`
}

type ProductResponse struct {
	Data Product `json:"data"`
	Meta Meta    `json:"meta"`
}

type ProductsResponse struct {
	Data []Product `json:"data"`
	Meta Meta      `json:"meta"`
}

type ProductImage struct {
	ID           int    `json:"id,omitempty"`
	ProductID    int    `json:"product_id,omitempty"`
	IsThumbnail  bool   `json:"is_thumbnail,omitempty"`
	SortOrder    int    `json:"sort_order,omitempty"`
	Description  string `json:"description,omitempty"`
	ImageFile    string `json:"image_file,omitempty"`
	URLZoom      string `json:"url_zoom,omitempty"`
	URLStandard  string `json:"url_standard,omitempty"`
	URLThumbnail string `json:"url_thumbnail,omitempty"`
	URLTiny      string `json:"url_tiny,omitempty"`
	DateModified string `json:"date_modified,omitempty"`
}

type ProductVideo struct {
	ID          int    `json:"id,omitempty"`
	ProductID   int    `json:"product_id,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	SortOrder   int    `json:"sort_order,omitempty"`
	Type        string `json:"type,omitempty"`
	VideoID     string `json:"video_id,omitempty"`
	URL         string `json:"url,omitempty"`
}

type ProductImageResponse struct {
	Data ProductImage `json:"data"`
	Meta Meta         `json:"meta"`
}

type ProductImagesResponse struct {
	Data []ProductImage `json:"data"`
	Meta Meta           `json:"meta"`
}

type ProductVideoResponse struct {
	Data ProductVideo `json:"data"`
	Meta Meta         `json:"meta"`
}

type ProductVideosResponse struct {
	Data []ProductVideo `json:"data"`
	Meta Meta           `json:"meta"`
}

type CustomURL struct {
	URL          string `json:"url,omitempty"`
	IsCustomized bool   `json:"is_customized,omitempty"`
}

type CustomField struct {
	ID        int    `json:"id,omitempty"`
	Name      string `json:"name"`
	Value     string `json:"value"`
	ProductID int    `json:"product_id,omitempty"`
}

type PricingRule struct {
	ID          int     `json:"id,omitempty"`
	QuantityMin int     `json:"quantity_min"`
	QuantityMax int     `json:"quantity_max,omitempty"`
	Type        string  `json:"type"`
	Amount      float64 `json:"amount"`
	ProductID   int     `json:"product_id,omitempty"`
}

type Variant struct {
	ID                     int           `json:"id,omitempty"`
	ProductID              int           `json:"product_id,omitempty"`
	SKU                    string        `json:"sku,omitempty"`
	Price                  float64       `json:"price,omitempty"`
	CostPrice              float64       `json:"cost_price,omitempty"`
	SalePrice              float64       `json:"sale_price,omitempty"`
	RetailPrice            float64       `json:"retail_price,omitempty"`
	Weight                 float64       `json:"weight,omitempty"`
	Width                  float64       `json:"width,omitempty"`
	Height                 float64       `json:"height,omitempty"`
	Depth                  float64       `json:"depth,omitempty"`
	IsFree                 bool          `json:"is_free_shipping,omitempty"`
	FixedCostShippingPrice float64       `json:"fixed_cost_shipping_price,omitempty"`
	PurchasingDisabled     bool          `json:"purchasing_disabled,omitempty"`
	PurchasingDisabledMsg  string        `json:"purchasing_disabled_message,omitempty"`
	ImageURL               string        `json:"image_url,omitempty"`
	UPC                    string        `json:"upc,omitempty"`
	MPN                    string        `json:"mpn,omitempty"`
	GTIN                   string        `json:"gtin,omitempty"`
	InventoryLevel         int           `json:"inventory_level,omitempty"`
	InventoryWarningLevel  int           `json:"inventory_warning_level,omitempty"`
	BinPickingNumber       string        `json:"bin_picking_number,omitempty"`
	OptionValues           []OptionValue `json:"option_values"`
}

type OptionValue struct {
	ID        int    `json:"id,omitempty"`
	OptionID  int    `json:"option_id"`
	Label     string `json:"label,omitempty"`
	SortOrder int    `json:"sort_order,omitempty"`
	Value     string `json:"value,omitempty"`
	IsDefault bool   `json:"is_default,omitempty"`
}

type ProductOption struct {
	ID           int           `json:"id,omitempty"`
	ProductID    int           `json:"product_id,omitempty"`
	DisplayName  string        `json:"display_name"`
	Type         string        `json:"type"`
	Config       OptionConfig  `json:"config,omitempty"`
	OptionValues []OptionValue `json:"option_values,omitempty"`
}

type OptionConfig struct {
	DefaultValue      string   `json:"default_value,omitempty"`
	CheckedByDefault  bool     `json:"checked_by_default,omitempty"`
	CheckboxLabel     string   `json:"checkbox_label,omitempty"`
	DateLimited       bool     `json:"date_limited,omitempty"`
	DateLimitMode     string   `json:"date_limit_mode,omitempty"`
	DateEarliestValue string   `json:"date_earliest_value,omitempty"`
	DateLatestValue   string   `json:"date_latest_value,omitempty"`
	FileTypes         []string `json:"file_types_supported,omitempty"`
	FileMaxSize       int      `json:"file_max_size,omitempty"`
	TextMinLength     int      `json:"text_min_length,omitempty"`
	TextMaxLength     int      `json:"text_max_length,omitempty"`
	TextCharsLimit    int      `json:"text_characters_limited,omitempty"`
	NumberLimited     bool     `json:"number_limited,omitempty"`
	NumberLimitMode   string   `json:"number_limit_mode,omitempty"`
	NumberMin         float64  `json:"number_lowest_value,omitempty"`
	NumberMax         float64  `json:"number_highest_value,omitempty"`
	NumberIntOnly     bool     `json:"number_integers_only,omitempty"`
}

type Modifier struct {
	ID           int            `json:"id,omitempty"`
	ProductID    int            `json:"product_id,omitempty"`
	Name         string         `json:"name"`
	DisplayName  string         `json:"display_name"`
	Type         string         `json:"type"`
	Required     bool           `json:"required"`
	Config       ModifierConfig `json:"config,omitempty"`
	OptionValues []OptionValue  `json:"option_values,omitempty"`
}

type ModifierConfig struct {
	DefaultValue          string   `json:"default_value,omitempty"`
	CheckedByDefault      bool     `json:"checked_by_default,omitempty"`
	CheckboxLabel         string   `json:"checkbox_label,omitempty"`
	DateLimited           bool     `json:"date_limited,omitempty"`
	DateLimitMode         string   `json:"date_limit_mode,omitempty"`
	DateEarliestValue     string   `json:"date_earliest_value,omitempty"`
	DateLatestValue       string   `json:"date_latest_value,omitempty"`
	FileTypes             []string `json:"file_types_supported,omitempty"`
	FileMaxSize           int      `json:"file_max_size,omitempty"`
	TextMinLength         int      `json:"text_min_length,omitempty"`
	TextMaxLength         int      `json:"text_max_length,omitempty"`
	TextCharsLimit        int      `json:"text_characters_limited,omitempty"`
	NumberLimited         bool     `json:"number_limited,omitempty"`
	NumberLimitMode       string   `json:"number_limit_mode,omitempty"`
	NumberMin             float64  `json:"number_lowest_value,omitempty"`
	NumberMax             float64  `json:"number_highest_value,omitempty"`
	NumberIntOnly         bool     `json:"number_integers_only,omitempty"`
	ProductListAdjuster   string   `json:"product_list_adjusts_inventory,omitempty"`
	ProductListAdjustName string   `json:"product_list_adjusts_pricing,omitempty"`
}

type Category struct {
	ID                 int        `json:"id,omitempty"`
	ParentID           int        `json:"parent_id"`
	Name               string     `json:"name"`
	Description        string     `json:"description,omitempty"`
	Views              int        `json:"views,omitempty"`
	SortOrder          int        `json:"sort_order,omitempty"`
	PageTitle          string     `json:"page_title,omitempty"`
	MetaKeywords       []string   `json:"meta_keywords,omitempty"`
	MetaDescription    string     `json:"meta_description,omitempty"`
	LayoutFile         string     `json:"layout_file,omitempty"`
	IsVisible          bool       `json:"is_visible,omitempty"`
	DefaultProductSort string     `json:"default_product_sort,omitempty"`
	ImageURL           string     `json:"image_url,omitempty"`
	CustomURL          *CustomURL `json:"custom_url,omitempty"`
}

type Brand struct {
	ID              int        `json:"id,omitempty"`
	Name            string     `json:"name"`
	PageTitle       string     `json:"page_title,omitempty"`
	MetaKeywords    []string   `json:"meta_keywords,omitempty"`
	MetaDescription string     `json:"meta_description,omitempty"`
	ImageURL        string     `json:"image_url,omitempty"`
	SearchKeywords  string     `json:"search_keywords,omitempty"`
	CustomURL       *CustomURL `json:"custom_url,omitempty"`
}

type Review struct {
	ID           int    `json:"id,omitempty"`
	ProductID    int    `json:"product_id,omitempty"`
	Title        string `json:"title"`
	Text         string `json:"text"`
	Status       string `json:"status,omitempty"`
	Rating       int    `json:"rating"`
	Email        string `json:"email,omitempty"`
	Name         string `json:"name,omitempty"`
	DateCreated  string `json:"date_created,omitempty"`
	DateModified string `json:"date_modified,omitempty"`
}

type ComplexRule struct {
	ID            int             `json:"id,omitempty"`
	ProductID     int             `json:"product_id,omitempty"`
	Enabled       bool            `json:"enabled"`
	Stop          bool            `json:"stop"`
	Purchasing    bool            `json:"purchasing_disabled"`
	PurchasingMsg string          `json:"purchasing_disabled_message,omitempty"`
	Adjusters     RuleAdjusters   `json:"price_adjuster,omitempty"`
	Conditions    []RuleCondition `json:"conditions"`
	SortOrder     int             `json:"sort_order,omitempty"`
}

type RuleAdjusters struct {
	Type   string  `json:"adjuster"`
	Amount float64 `json:"adjuster_value,omitempty"`
}

type RuleCondition struct {
	OptionID int    `json:"product_option_id"`
	ValueID  int    `json:"product_option_value_id"`
	Rule     string `json:"rule,omitempty"`
}

type Channel struct {
	ID                  int    `json:"id,omitempty"`
	Name                string `json:"name"`
	Type                string `json:"type"`
	Platform            string `json:"platform,omitempty"`
	Status              string `json:"status"`
	IsListable          bool   `json:"is_listable"`
	IsVisible           bool   `json:"is_visible"`
	ExternalID          string `json:"external_id,omitempty"`
	IsExternallyManaged bool   `json:"is_externally_managed,omitempty"`
	DateCreated         string `json:"date_created,omitempty"`
	DateModified        string `json:"date_modified,omitempty"`
}

type Metafield struct {
	ID           int    `json:"id,omitempty"`
	Key          string `json:"key"`
	Value        string `json:"value"`
	ResourceType string `json:"resource_type"`
	ResourceID   int    `json:"resource_id"`
	Description  string `json:"description,omitempty"`
	Namespace    string `json:"namespace"`
	Permission   string `json:"permission,omitempty"`
	DateCreated  string `json:"date_created,omitempty"`
	DateModified string `json:"date_modified,omitempty"`
}

type Summary struct {
	InventoryLevel   int     `json:"inventory_level"`
	InventoryWarning int     `json:"inventory_warning_level"`
	PrimaryCategory  int     `json:"primary_category_id"`
	TotalSold        int     `json:"total_sold"`
	PrimaryImage     *Image  `json:"primary_image,omitempty"`
	Availability     string  `json:"availability"`
	AvgRating        float64 `json:"rating_average"`
	NumReviews       int     `json:"number_of_reviews"`
}

type Image struct {
	ID           int    `json:"id,omitempty"`
	ProductID    int    `json:"product_id,omitempty"`
	URLThumbnail string `json:"url_thumbnail,omitempty"`
	URLStandard  string `json:"url_standard,omitempty"`
}

type CategoryResponse struct {
	Data Category `json:"data"`
	Meta Meta     `json:"meta"`
}

type CategoriesResponse struct {
	Data []Category `json:"data"`
	Meta Meta       `json:"meta"`
}

type BrandResponse struct {
	Data Brand `json:"data"`
	Meta Meta  `json:"meta"`
}

type BrandsResponse struct {
	Data []Brand `json:"data"`
	Meta Meta    `json:"meta"`
}

type VariantResponse struct {
	Data Variant `json:"data"`
	Meta Meta    `json:"meta"`
}

type VariantsResponse struct {
	Data []Variant `json:"data"`
	Meta Meta      `json:"meta"`
}

type OptionValueResponse struct {
	Data OptionValue `json:"data"`
	Meta Meta        `json:"meta"`
}

type OptionValuesResponse struct {
	Data []OptionValue `json:"data"`
	Meta Meta          `json:"meta"`
}

type ProductOptionResponse struct {
	Data ProductOption `json:"data"`
	Meta Meta          `json:"meta"`
}

type ProductOptionsResponse struct {
	Data []ProductOption `json:"data"`
	Meta Meta            `json:"meta"`
}

type ModifierResponse struct {
	Data Modifier `json:"data"`
	Meta Meta     `json:"meta"`
}

type ModifiersResponse struct {
	Data []Modifier `json:"data"`
	Meta Meta       `json:"meta"`
}

type ReviewResponse struct {
	Data Review `json:"data"`
	Meta Meta   `json:"meta"`
}

type ReviewsResponse struct {
	Data []Review `json:"data"`
	Meta Meta     `json:"meta"`
}

type ComplexRuleResponse struct {
	Data ComplexRule `json:"data"`
	Meta Meta        `json:"meta"`
}

type ComplexRulesResponse struct {
	Data []ComplexRule `json:"data"`
	Meta Meta          `json:"meta"`
}

type CustomFieldResponse struct {
	Data CustomField `json:"data"`
	Meta Meta        `json:"meta"`
}

type CustomFieldsResponse struct {
	Data []CustomField `json:"data"`
	Meta Meta          `json:"meta"`
}

type PricingRuleResponse struct {
	Data PricingRule `json:"data"`
	Meta Meta        `json:"meta"`
}

type PricingRulesResponse struct {
	Data []PricingRule `json:"data"`
	Meta Meta          `json:"meta"`
}

type ChannelResponse struct {
	Data Channel `json:"data"`
	Meta Meta    `json:"meta"`
}

type ChannelsResponse struct {
	Data []Channel `json:"data"`
	Meta Meta      `json:"meta"`
}

type MetafieldResponse struct {
	Data Metafield `json:"data"`
	Meta Meta      `json:"meta"`
}

type MetafieldsResponse struct {
	Data []Metafield `json:"data"`
	Meta Meta        `json:"meta"`
}

type SummaryResponse struct {
	Data Summary `json:"data"`
	Meta Meta    `json:"meta"`
}

type BrandsService struct {
	client *Client
}

func (s *BrandsService) ListContext(ctx context.Context, params *QueryParams) (*BrandsResponse, error) {
	path := "catalog/brands"

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	if params != nil {
		req.URL.RawQuery = params.ToValues().Encode()
	}

	brandsResponse := new(BrandsResponse)
	_, err = s.client.Do(req, brandsResponse)
	return brandsResponse, err
}

func (s *BrandsService) GetContext(ctx context.Context, id int, params *QueryParams) (*BrandResponse, error) {
	path := fmt.Sprintf("catalog/brands/%d", id)

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	if params != nil {
		req.URL.RawQuery = params.ToValues().Encode()
	}

	brandResponse := new(BrandResponse)
	_, err = s.client.Do(req, brandResponse)
	return brandResponse, err
}

func (s *BrandsService) CreateContext(ctx context.Context, brand *Brand) (*BrandResponse, error) {
	path := "catalog/brands"

	req, err := s.client.NewRequest(ctx, "POST", path, brand)
	if err != nil {
		return nil, err
	}

	brandResponse := new(BrandResponse)
	_, err = s.client.Do(req, brandResponse)
	return brandResponse, err
}

func (s *BrandsService) UpdateContext(ctx context.Context, id int, brand *Brand) (*BrandResponse, error) {
	path := fmt.Sprintf("catalog/brands/%d", id)

	req, err := s.client.NewRequest(ctx, "PUT", path, brand)
	if err != nil {
		return nil, err
	}

	brandResponse := new(BrandResponse)
	_, err = s.client.Do(req, brandResponse)
	return brandResponse, err
}

func (s *BrandsService) DeleteContext(ctx context.Context, id int) error {
	path := fmt.Sprintf("catalog/brands/%d", id)

	req, err := s.client.NewRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(req, nil)
	return err
}

type CategoriesService struct {
	client *Client
}

func (s *CategoriesService) ListContext(ctx context.Context, params *QueryParams) (*CategoriesResponse, error) {
	path := "catalog/categories"

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	if params != nil {
		req.URL.RawQuery = params.ToValues().Encode()
	}

	categoriesResponse := new(CategoriesResponse)
	_, err = s.client.Do(req, categoriesResponse)
	return categoriesResponse, err
}

func (s *CategoriesService) GetContext(ctx context.Context, id int, params *QueryParams) (*CategoryResponse, error) {
	path := fmt.Sprintf("catalog/categories/%d", id)

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	if params != nil {
		req.URL.RawQuery = params.ToValues().Encode()
	}

	categoryResponse := new(CategoryResponse)
	_, err = s.client.Do(req, categoryResponse)
	return categoryResponse, err
}

func (s *CategoriesService) CreateContext(ctx context.Context, category *Category) (*CategoryResponse, error) {
	path := "catalog/categories"

	req, err := s.client.NewRequest(ctx, "POST", path, category)
	if err != nil {
		return nil, err
	}

	categoryResponse := new(CategoryResponse)
	_, err = s.client.Do(req, categoryResponse)
	return categoryResponse, err
}

func (s *CategoriesService) UpdateContext(ctx context.Context, id int, category *Category) (*CategoryResponse, error) {
	path := fmt.Sprintf("catalog/categories/%d", id)

	req, err := s.client.NewRequest(ctx, "PUT", path, category)
	if err != nil {
		return nil, err
	}

	categoryResponse := new(CategoryResponse)
	_, err = s.client.Do(req, categoryResponse)
	return categoryResponse, err
}

func (s *CategoriesService) DeleteContext(ctx context.Context, id int) error {
	path := fmt.Sprintf("catalog/categories/%d", id)

	req, err := s.client.NewRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(req, nil)
	return err
}

type ChannelsService struct {
	client *Client
}

func (s *ChannelsService) ListContext(ctx context.Context, params *QueryParams) (*ChannelsResponse, error) {
	path := "channels"

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	if params != nil {
		req.URL.RawQuery = params.ToValues().Encode()
	}

	channelsResponse := new(ChannelsResponse)
	_, err = s.client.Do(req, channelsResponse)
	return channelsResponse, err
}

func (s *ChannelsService) GetContext(ctx context.Context, channelID int) (*ChannelResponse, error) {
	path := fmt.Sprintf("channels/%d", channelID)

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	channelResponse := new(ChannelResponse)
	_, err = s.client.Do(req, channelResponse)
	return channelResponse, err
}

func (s *ChannelsService) CreateContext(ctx context.Context, channel *Channel) (*ChannelResponse, error) {
	path := "channels"

	req, err := s.client.NewRequest(ctx, "POST", path, channel)
	if err != nil {
		return nil, err
	}

	channelResponse := new(ChannelResponse)
	_, err = s.client.Do(req, channelResponse)
	return channelResponse, err
}

func (s *ChannelsService) UpdateContext(ctx context.Context, channelID int, channel *Channel) (*ChannelResponse, error) {
	path := fmt.Sprintf("channels/%d", channelID)

	req, err := s.client.NewRequest(ctx, "PUT", path, channel)
	if err != nil {
		return nil, err
	}

	channelResponse := new(ChannelResponse)
	_, err = s.client.Do(req, channelResponse)
	return channelResponse, err
}

func (s *ChannelsService) DeleteContext(ctx context.Context, channelID int) error {
	path := fmt.Sprintf("channels/%d", channelID)

	req, err := s.client.NewRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(req, nil)
	return err
}

type ComplexRulesService struct {
	client *Client
}

func (s *ComplexRulesService) ListContext(ctx context.Context, productID int, params *QueryParams) (*ComplexRulesResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/complex-rules", productID)

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	if params != nil {
		req.URL.RawQuery = params.ToValues().Encode()
	}

	rulesResponse := new(ComplexRulesResponse)
	_, err = s.client.Do(req, rulesResponse)
	return rulesResponse, err
}

func (s *ComplexRulesService) GetContext(ctx context.Context, productID, ruleID int) (*ComplexRuleResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/complex-rules/%d", productID, ruleID)

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	ruleResponse := new(ComplexRuleResponse)
	_, err = s.client.Do(req, ruleResponse)
	return ruleResponse, err
}

func (s *ComplexRulesService) CreateContext(ctx context.Context, productID int, rule *ComplexRule) (*ComplexRuleResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/complex-rules", productID)

	req, err := s.client.NewRequest(ctx, "POST", path, rule)
	if err != nil {
		return nil, err
	}

	ruleResponse := new(ComplexRuleResponse)
	_, err = s.client.Do(req, ruleResponse)
	return ruleResponse, err
}

func (s *ComplexRulesService) UpdateContext(ctx context.Context, productID, ruleID int, rule *ComplexRule) (*ComplexRuleResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/complex-rules/%d", productID, ruleID)

	req, err := s.client.NewRequest(ctx, "PUT", path, rule)
	if err != nil {
		return nil, err
	}

	ruleResponse := new(ComplexRuleResponse)
	_, err = s.client.Do(req, ruleResponse)
	return ruleResponse, err
}

func (s *ComplexRulesService) DeleteContext(ctx context.Context, productID, ruleID int) error {
	path := fmt.Sprintf("catalog/products/%d/complex-rules/%d", productID, ruleID)

	req, err := s.client.NewRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(req, nil)
	return err
}

type CustomFieldsService struct {
	client *Client
}

func (s *CustomFieldsService) ListContext(ctx context.Context, productID int, params *QueryParams) (*CustomFieldsResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/custom-fields", productID)

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	if params != nil {
		req.URL.RawQuery = params.ToValues().Encode()
	}

	fieldsResponse := new(CustomFieldsResponse)
	_, err = s.client.Do(req, fieldsResponse)
	return fieldsResponse, err
}

func (s *CustomFieldsService) GetContext(ctx context.Context, productID, fieldID int) (*CustomFieldResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/custom-fields/%d", productID, fieldID)

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	fieldResponse := new(CustomFieldResponse)
	_, err = s.client.Do(req, fieldResponse)
	return fieldResponse, err
}

func (s *CustomFieldsService) CreateContext(ctx context.Context, productID int, field *CustomField) (*CustomFieldResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/custom-fields", productID)

	req, err := s.client.NewRequest(ctx, "POST", path, field)
	if err != nil {
		return nil, err
	}

	fieldResponse := new(CustomFieldResponse)
	_, err = s.client.Do(req, fieldResponse)
	return fieldResponse, err
}

func (s *CustomFieldsService) UpdateContext(ctx context.Context, productID, fieldID int, field *CustomField) (*CustomFieldResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/custom-fields/%d", productID, fieldID)

	req, err := s.client.NewRequest(ctx, "PUT", path, field)
	if err != nil {
		return nil, err
	}

	fieldResponse := new(CustomFieldResponse)
	_, err = s.client.Do(req, fieldResponse)
	return fieldResponse, err
}

func (s *CustomFieldsService) DeleteContext(ctx context.Context, productID, fieldID int) error {
	path := fmt.Sprintf("catalog/products/%d/custom-fields/%d", productID, fieldID)

	req, err := s.client.NewRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(req, nil)
	return err
}

type ProductImagesService struct {
	client *Client
}

func (s *ProductImagesService) ListContext(ctx context.Context, productID int, params *QueryParams) (*ProductImagesResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/images", productID)

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	if params != nil {
		req.URL.RawQuery = params.ToValues().Encode()
	}

	imagesResponse := new(ProductImagesResponse)
	_, err = s.client.Do(req, imagesResponse)
	return imagesResponse, err
}

func (s *ProductImagesService) GetContext(ctx context.Context, productID, imageID int) (*ProductImageResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/images/%d", productID, imageID)

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	imageResponse := new(ProductImageResponse)
	_, err = s.client.Do(req, imageResponse)
	return imageResponse, err
}

func (s *ProductImagesService) CreateContext(ctx context.Context, productID int, image *ProductImage) (*ProductImageResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/images", productID)

	req, err := s.client.NewRequest(ctx, "POST", path, image)
	if err != nil {
		return nil, err
	}

	imageResponse := new(ProductImageResponse)
	_, err = s.client.Do(req, imageResponse)
	return imageResponse, err
}

func (s *ProductImagesService) UpdateContext(ctx context.Context, productID, imageID int, image *ProductImage) (*ProductImageResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/images/%d", productID, imageID)

	req, err := s.client.NewRequest(ctx, "PUT", path, image)
	if err != nil {
		return nil, err
	}

	imageResponse := new(ProductImageResponse)
	_, err = s.client.Do(req, imageResponse)
	return imageResponse, err
}

func (s *ProductImagesService) DeleteContext(ctx context.Context, productID, imageID int) error {
	path := fmt.Sprintf("catalog/products/%d/images/%d", productID, imageID)

	req, err := s.client.NewRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(req, nil)
	return err
}

type MetafieldsService struct {
	client *Client
}

func (s *MetafieldsService) ListContext(ctx context.Context, resourceType string, resourceID int, params *QueryParams) (*MetafieldsResponse, error) {
	path := fmt.Sprintf("catalog/%s/%d/metafields", resourceType, resourceID)

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	if params != nil {
		req.URL.RawQuery = params.ToValues().Encode()
	}

	metafieldsResponse := new(MetafieldsResponse)
	_, err = s.client.Do(req, metafieldsResponse)
	return metafieldsResponse, err
}

func (s *MetafieldsService) GetContext(ctx context.Context, resourceType string, resourceID, metafieldID int) (*MetafieldResponse, error) {
	path := fmt.Sprintf("catalog/%s/%d/metafields/%d", resourceType, resourceID, metafieldID)

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	metafieldResponse := new(MetafieldResponse)
	_, err = s.client.Do(req, metafieldResponse)
	return metafieldResponse, err
}

func (s *MetafieldsService) CreateContext(ctx context.Context, resourceType string, resourceID int, metafield *Metafield) (*MetafieldResponse, error) {
	path := fmt.Sprintf("catalog/%s/%d/metafields", resourceType, resourceID)

	req, err := s.client.NewRequest(ctx, "POST", path, metafield)
	if err != nil {
		return nil, err
	}

	metafieldResponse := new(MetafieldResponse)
	_, err = s.client.Do(req, metafieldResponse)
	return metafieldResponse, err
}

func (s *MetafieldsService) UpdateContext(ctx context.Context, resourceType string, resourceID, metafieldID int, metafield *Metafield) (*MetafieldResponse, error) {
	path := fmt.Sprintf("catalog/%s/%d/metafields/%d", resourceType, resourceID, metafieldID)

	req, err := s.client.NewRequest(ctx, "PUT", path, metafield)
	if err != nil {
		return nil, err
	}

	metafieldResponse := new(MetafieldResponse)
	_, err = s.client.Do(req, metafieldResponse)
	return metafieldResponse, err
}

func (s *MetafieldsService) DeleteContext(ctx context.Context, resourceType string, resourceID, metafieldID int) error {
	path := fmt.Sprintf("catalog/%s/%d/metafields/%d", resourceType, resourceID, metafieldID)

	req, err := s.client.NewRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(req, nil)
	return err
}

type ModifiersService struct {
	client *Client
}

func (s *ModifiersService) ListContext(ctx context.Context, productID int, params *QueryParams) (*ModifiersResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/modifiers", productID)

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	if params != nil {
		req.URL.RawQuery = params.ToValues().Encode()
	}

	modifiersResponse := new(ModifiersResponse)
	_, err = s.client.Do(req, modifiersResponse)
	return modifiersResponse, err
}

func (s *ModifiersService) GetContext(ctx context.Context, productID, modifierID int) (*ModifierResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/modifiers/%d", productID, modifierID)

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	modifierResponse := new(ModifierResponse)
	_, err = s.client.Do(req, modifierResponse)
	return modifierResponse, err
}

func (s *ModifiersService) CreateContext(ctx context.Context, productID int, modifier *Modifier) (*ModifierResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/modifiers", productID)

	req, err := s.client.NewRequest(ctx, "POST", path, modifier)
	if err != nil {
		return nil, err
	}

	modifierResponse := new(ModifierResponse)
	_, err = s.client.Do(req, modifierResponse)
	return modifierResponse, err
}

func (s *ModifiersService) UpdateContext(ctx context.Context, productID, modifierID int, modifier *Modifier) (*ModifierResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/modifiers/%d", productID, modifierID)

	req, err := s.client.NewRequest(ctx, "PUT", path, modifier)
	if err != nil {
		return nil, err
	}

	modifierResponse := new(ModifierResponse)
	_, err = s.client.Do(req, modifierResponse)
	return modifierResponse, err
}

func (s *ModifiersService) DeleteContext(ctx context.Context, productID, modifierID int) error {
	path := fmt.Sprintf("catalog/products/%d/modifiers/%d", productID, modifierID)

	req, err := s.client.NewRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(req, nil)
	return err
}

func (s *ModifiersService) GetModifierValuesContext(ctx context.Context, productID, modifierID int, params *QueryParams) (*OptionValuesResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/modifiers/%d/values", productID, modifierID)

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	if params != nil {
		req.URL.RawQuery = params.ToValues().Encode()
	}

	valuesResponse := new(OptionValuesResponse)
	_, err = s.client.Do(req, valuesResponse)
	return valuesResponse, err
}

func (s *ModifiersService) GetModifierValueContext(ctx context.Context, productID, modifierID, valueID int) (*OptionValueResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/modifiers/%d/values/%d", productID, modifierID, valueID)

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	valueResponse := new(OptionValueResponse)
	_, err = s.client.Do(req, valueResponse)
	return valueResponse, err
}

func (s *ModifiersService) CreateModifierValueContext(ctx context.Context, productID, modifierID int, value *OptionValue) (*OptionValueResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/modifiers/%d/values", productID, modifierID)

	req, err := s.client.NewRequest(ctx, "POST", path, value)
	if err != nil {
		return nil, err
	}

	valueResponse := new(OptionValueResponse)
	_, err = s.client.Do(req, valueResponse)
	return valueResponse, err
}

func (s *ModifiersService) UpdateModifierValueContext(ctx context.Context, productID, modifierID, valueID int, value *OptionValue) (*OptionValueResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/modifiers/%d/values/%d", productID, modifierID, valueID)

	req, err := s.client.NewRequest(ctx, "PUT", path, value)
	if err != nil {
		return nil, err
	}

	valueResponse := new(OptionValueResponse)
	_, err = s.client.Do(req, valueResponse)
	return valueResponse, err
}

func (s *ModifiersService) DeleteModifierValueContext(ctx context.Context, productID, modifierID, valueID int) error {
	path := fmt.Sprintf("catalog/products/%d/modifiers/%d/values/%d", productID, modifierID, valueID)

	req, err := s.client.NewRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(req, nil)
	return err
}

type OptionsService struct {
	client *Client
}

func (s *OptionsService) ListContext(ctx context.Context, productID int, params *QueryParams) (*ProductOptionsResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/options", productID)

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	if params != nil {
		req.URL.RawQuery = params.ToValues().Encode()
	}

	optionsResponse := new(ProductOptionsResponse)
	_, err = s.client.Do(req, optionsResponse)
	return optionsResponse, err
}

func (s *OptionsService) GetContext(ctx context.Context, productID, optionID int) (*ProductOptionResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/options/%d", productID, optionID)

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	optionResponse := new(ProductOptionResponse)
	_, err = s.client.Do(req, optionResponse)
	return optionResponse, err
}

func (s *OptionsService) CreateContext(ctx context.Context, productID int, option *ProductOption) (*ProductOptionResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/options", productID)

	req, err := s.client.NewRequest(ctx, "POST", path, option)
	if err != nil {
		return nil, err
	}

	optionResponse := new(ProductOptionResponse)
	_, err = s.client.Do(req, optionResponse)
	return optionResponse, err
}

func (s *OptionsService) UpdateContext(ctx context.Context, productID, optionID int, option *ProductOption) (*ProductOptionResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/options/%d", productID, optionID)

	req, err := s.client.NewRequest(ctx, "PUT", path, option)
	if err != nil {
		return nil, err
	}

	optionResponse := new(ProductOptionResponse)
	_, err = s.client.Do(req, optionResponse)
	return optionResponse, err
}

func (s *OptionsService) DeleteContext(ctx context.Context, productID, optionID int) error {
	path := fmt.Sprintf("catalog/products/%d/options/%d", productID, optionID)

	req, err := s.client.NewRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(req, nil)
	return err
}

func (s *OptionsService) GetOptionValuesContext(ctx context.Context, productID, optionID int, params *QueryParams) (*OptionValuesResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/options/%d/values", productID, optionID)

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	if params != nil {
		req.URL.RawQuery = params.ToValues().Encode()
	}

	valuesResponse := new(OptionValuesResponse)
	_, err = s.client.Do(req, valuesResponse)
	return valuesResponse, err
}

func (s *OptionsService) GetOptionValueContext(ctx context.Context, productID, optionID, valueID int) (*OptionValueResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/options/%d/values/%d", productID, optionID, valueID)

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	valueResponse := new(OptionValueResponse)
	_, err = s.client.Do(req, valueResponse)
	return valueResponse, err
}

func (s *OptionsService) CreateOptionValueContext(ctx context.Context, productID, optionID int, value *OptionValue) (*OptionValueResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/options/%d/values", productID, optionID)

	req, err := s.client.NewRequest(ctx, "POST", path, value)
	if err != nil {
		return nil, err
	}

	valueResponse := new(OptionValueResponse)
	_, err = s.client.Do(req, valueResponse)
	return valueResponse, err
}

func (s *OptionsService) UpdateOptionValueContext(ctx context.Context, productID, optionID, valueID int, value *OptionValue) (*OptionValueResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/options/%d/values/%d", productID, optionID, valueID)

	req, err := s.client.NewRequest(ctx, "PUT", path, value)
	if err != nil {
		return nil, err
	}

	valueResponse := new(OptionValueResponse)
	_, err = s.client.Do(req, valueResponse)
	return valueResponse, err
}

func (s *OptionsService) DeleteOptionValueContext(ctx context.Context, productID, optionID, valueID int) error {
	path := fmt.Sprintf("catalog/products/%d/options/%d/values/%d", productID, optionID, valueID)

	req, err := s.client.NewRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(req, nil)
	return err
}

type ProductsService struct {
	client *Client
}

func (s *ProductsService) ListContext(ctx context.Context, params *QueryParams) (*ProductsResponse, error) {
	path := "catalog/products"

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	if params != nil {
		req.URL.RawQuery = params.ToValues().Encode()
	}

	productsResponse := new(ProductsResponse)
	_, err = s.client.Do(req, productsResponse)
	return productsResponse, err
}

func (s *ProductsService) GetContext(ctx context.Context, id int, params *QueryParams) (*ProductResponse, error) {
	path := fmt.Sprintf("catalog/products/%d", id)

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	if params != nil {
		req.URL.RawQuery = params.ToValues().Encode()
	}

	productResponse := new(ProductResponse)
	_, err = s.client.Do(req, productResponse)
	return productResponse, err
}

func (s *ProductsService) CreateContext(ctx context.Context, product *Product) (*ProductResponse, error) {
	path := "catalog/products"

	req, err := s.client.NewRequest(ctx, "POST", path, product)
	if err != nil {
		return nil, err
	}

	productResponse := new(ProductResponse)
	_, err = s.client.Do(req, productResponse)
	return productResponse, err
}

func (s *ProductsService) UpdateContext(ctx context.Context, id int, product *Product) (*ProductResponse, error) {
	path := fmt.Sprintf("catalog/products/%d", id)

	req, err := s.client.NewRequest(ctx, "PUT", path, product)
	if err != nil {
		return nil, err
	}

	productResponse := new(ProductResponse)
	_, err = s.client.Do(req, productResponse)
	return productResponse, err
}

func (s *ProductsService) DeleteContext(ctx context.Context, id int) error {
	path := fmt.Sprintf("catalog/products/%d", id)

	req, err := s.client.NewRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(req, nil)
	return err
}

type ReviewsService struct {
	client *Client
}

func (s *ReviewsService) ListContext(ctx context.Context, productID int, params *QueryParams) (*ReviewsResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/reviews", productID)

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	if params != nil {
		req.URL.RawQuery = params.ToValues().Encode()
	}

	reviewsResponse := new(ReviewsResponse)
	_, err = s.client.Do(req, reviewsResponse)
	return reviewsResponse, err
}

func (s *ReviewsService) GetContext(ctx context.Context, productID, reviewID int) (*ReviewResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/reviews/%d", productID, reviewID)

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	reviewResponse := new(ReviewResponse)
	_, err = s.client.Do(req, reviewResponse)
	return reviewResponse, err
}

func (s *ReviewsService) CreateContext(ctx context.Context, productID int, review *Review) (*ReviewResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/reviews", productID)

	req, err := s.client.NewRequest(ctx, "POST", path, review)
	if err != nil {
		return nil, err
	}

	reviewResponse := new(ReviewResponse)
	_, err = s.client.Do(req, reviewResponse)
	return reviewResponse, err
}

func (s *ReviewsService) UpdateContext(ctx context.Context, productID, reviewID int, review *Review) (*ReviewResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/reviews/%d", productID, reviewID)

	req, err := s.client.NewRequest(ctx, "PUT", path, review)
	if err != nil {
		return nil, err
	}

	reviewResponse := new(ReviewResponse)
	_, err = s.client.Do(req, reviewResponse)
	return reviewResponse, err
}

func (s *ReviewsService) DeleteContext(ctx context.Context, productID, reviewID int) error {
	path := fmt.Sprintf("catalog/products/%d/reviews/%d", productID, reviewID)

	req, err := s.client.NewRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(req, nil)
	return err
}

type SummaryService struct {
	client *Client
}

func (s *SummaryService) GetContext(ctx context.Context, productID int) (*SummaryResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/summary", productID)

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	summaryResponse := new(SummaryResponse)
	_, err = s.client.Do(req, summaryResponse)
	return summaryResponse, err
}

type VariantsService struct {
	client *Client
}

func (s *VariantsService) ListContext(ctx context.Context, productID int, params *QueryParams) (*VariantsResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/variants", productID)

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	if params != nil {
		req.URL.RawQuery = params.ToValues().Encode()
	}

	variantsResponse := new(VariantsResponse)
	_, err = s.client.Do(req, variantsResponse)
	return variantsResponse, err
}

func (s *VariantsService) GetContext(ctx context.Context, productID, variantID int) (*VariantResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/variants/%d", productID, variantID)

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	variantResponse := new(VariantResponse)
	_, err = s.client.Do(req, variantResponse)
	return variantResponse, err
}

func (s *VariantsService) CreateContext(ctx context.Context, productID int, variant *Variant) (*VariantResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/variants", productID)

	req, err := s.client.NewRequest(ctx, "POST", path, variant)
	if err != nil {
		return nil, err
	}

	variantResponse := new(VariantResponse)
	_, err = s.client.Do(req, variantResponse)
	return variantResponse, err
}

func (s *VariantsService) UpdateContext(ctx context.Context, productID, variantID int, variant *Variant) (*VariantResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/variants/%d", productID, variantID)

	req, err := s.client.NewRequest(ctx, "PUT", path, variant)
	if err != nil {
		return nil, err
	}

	variantResponse := new(VariantResponse)
	_, err = s.client.Do(req, variantResponse)
	return variantResponse, err
}

func (s *VariantsService) DeleteContext(ctx context.Context, productID, variantID int) error {
	path := fmt.Sprintf("catalog/products/%d/variants/%d", productID, variantID)

	req, err := s.client.NewRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(req, nil)
	return err
}

type VideosService struct {
	client *Client
}

func (s *VideosService) ListContext(ctx context.Context, productID int, params *QueryParams) (*ProductVideosResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/videos", productID)

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	if params != nil {
		req.URL.RawQuery = params.ToValues().Encode()
	}

	videosResponse := new(ProductVideosResponse)
	_, err = s.client.Do(req, videosResponse)
	return videosResponse, err
}

func (s *VideosService) GetContext(ctx context.Context, productID, videoID int) (*ProductVideoResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/videos/%d", productID, videoID)

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	videoResponse := new(ProductVideoResponse)
	_, err = s.client.Do(req, videoResponse)
	return videoResponse, err
}

func (s *VideosService) CreateContext(ctx context.Context, productID int, video *ProductVideo) (*ProductVideoResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/videos", productID)

	req, err := s.client.NewRequest(ctx, "POST", path, video)
	if err != nil {
		return nil, err
	}

	videoResponse := new(ProductVideoResponse)
	_, err = s.client.Do(req, videoResponse)
	return videoResponse, err
}

func (s *VideosService) UpdateContext(ctx context.Context, productID, videoID int, video *ProductVideo) (*ProductVideoResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/videos/%d", productID, videoID)

	req, err := s.client.NewRequest(ctx, "PUT", path, video)
	if err != nil {
		return nil, err
	}

	videoResponse := new(ProductVideoResponse)
	_, err = s.client.Do(req, videoResponse)
	return videoResponse, err
}

func (s *VideosService) DeleteContext(ctx context.Context, productID, videoID int) error {
	path := fmt.Sprintf("catalog/products/%d/videos/%d", productID, videoID)

	req, err := s.client.NewRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(req, nil)
	return err
}

type ProductChannelAssignment struct {
	ProductID int `json:"product_id"`
	ChannelID int `json:"channel_id"`
}

type ProductChannelAssignmentsResponse struct {
	Data []ProductChannelAssignment `json:"data"`
	Meta Meta                       `json:"meta"`
}

type CategoryAssignment struct {
	ProductID  int `json:"product_id"`
	CategoryID int `json:"category_id"`
}

type CategoryAssignmentsResponse struct {
	Data []CategoryAssignment `json:"data"`
	Meta Meta                 `json:"meta"`
}

type BatchErrorResponse struct {
	Data []BatchError `json:"data"`
	Meta Meta         `json:"meta"`
}

type BatchError struct {
	Error       string `json:"error"`
	ResourceID  int    `json:"resource_id,omitempty"`
	ResourceURL string `json:"resource_url,omitempty"`
	Status      int    `json:"status"`
}

type BatchProductDeleteRequest struct {
	IDs []int `json:"product_ids"`
}

type BatchProductCreateRequest struct {
	Products []Product `json:"products"`
}

type BatchProductUpdateRequest struct {
	Products []Product `json:"products"`
}

type BatchProductsResponse struct {
	Data []Product `json:"data"`
	Meta Meta      `json:"meta"`
}

type PricingRequest struct {
	ProductIDs   []int                      `json:"product_ids,omitempty"`
	VariantIDs   []int                      `json:"variant_ids,omitempty"`
	IncludeTaxes bool                       `json:"include_taxes,omitempty"`
	Currencies   []string                   `json:"currencies,omitempty"`
	Context      PricingRequestContext      `json:"context,omitempty"`
	Filters      PricingRequestFilters      `json:"filters,omitempty"`
	Aggregations PricingRequestAggregations `json:"aggregations,omitempty"`
}

type PricingRequestContext struct {
	ChannelID       int `json:"channel_id,omitempty"`
	CustomerID      int `json:"customer_id,omitempty"`
	CustomerGroupID int `json:"customer_group_id,omitempty"`
}

type PricingRequestFilters struct {
	SalesTaxable bool `json:"sales_taxable,omitempty"`
}

type PricingRequestAggregations struct {
	TaxExcludedPriceMin bool `json:"tax_excluded_price_min,omitempty"`
	TaxExcludedPriceMax bool `json:"tax_excluded_price_max,omitempty"`
	TaxIncludedPriceMin bool `json:"tax_included_price_min,omitempty"`
	TaxIncludedPriceMax bool `json:"tax_included_price_max,omitempty"`
}

type PricingResponse struct {
	Data PricingData `json:"data"`
	Meta Meta        `json:"meta"`
}

type PricingData struct {
	Products     map[string]PricingProductData `json:"products,omitempty"`
	Variants     map[string]PricingVariantData `json:"variants,omitempty"`
	Aggregations PricingAggregationData        `json:"aggregations,omitempty"`
}

type PricingProductData struct {
	PriceExcludingTax       float64       `json:"price_excluding_tax"`
	PriceIncludingTax       float64       `json:"price_including_tax"`
	TaxAmount               float64       `json:"tax_amount"`
	RetailPriceExcludingTax float64       `json:"retail_price_excluding_tax"`
	RetailPriceIncludingTax float64       `json:"retail_price_including_tax"`
	RetailTaxAmount         float64       `json:"retail_tax_amount"`
	SalePriceExcludingTax   float64       `json:"sale_price_excluding_tax"`
	SalePriceIncludingTax   float64       `json:"sale_price_including_tax"`
	SaleTaxAmount           float64       `json:"sale_tax_amount"`
	MapPriceExcludingTax    float64       `json:"map_price_excluding_tax"`
	MapPriceIncludingTax    float64       `json:"map_price_including_tax"`
	MapTaxAmount            float64       `json:"map_tax_amount"`
	BulkPricingTiers        []PricingTier `json:"bulk_pricing_tiers"`
	Currency                string        `json:"currency"`
}

type PricingVariantData struct {
	PriceExcludingTax       float64       `json:"price_excluding_tax"`
	PriceIncludingTax       float64       `json:"price_including_tax"`
	TaxAmount               float64       `json:"tax_amount"`
	RetailPriceExcludingTax float64       `json:"retail_price_excluding_tax"`
	RetailPriceIncludingTax float64       `json:"retail_price_including_tax"`
	RetailTaxAmount         float64       `json:"retail_tax_amount"`
	SalePriceExcludingTax   float64       `json:"sale_price_excluding_tax"`
	SalePriceIncludingTax   float64       `json:"sale_price_including_tax"`
	SaleTaxAmount           float64       `json:"sale_tax_amount"`
	MapPriceExcludingTax    float64       `json:"map_price_excluding_tax"`
	MapPriceIncludingTax    float64       `json:"map_price_including_tax"`
	MapTaxAmount            float64       `json:"map_tax_amount"`
	BulkPricingTiers        []PricingTier `json:"bulk_pricing_tiers"`
	Currency                string        `json:"currency"`
}

type PricingTier struct {
	QuantityMin       int     `json:"quantity_min"`
	QuantityMax       int     `json:"quantity_max,omitempty"`
	Type              string  `json:"type"`
	Amount            float64 `json:"amount"`
	PriceExcludingTax float64 `json:"price_excluding_tax"`
	PriceIncludingTax float64 `json:"price_including_tax"`
	TaxAmount         float64 `json:"tax_amount"`
}

type PricingAggregationData struct {
	TaxExcludedPriceMin float64 `json:"tax_excluded_price_min,omitempty"`
	TaxExcludedPriceMax float64 `json:"tax_excluded_price_max,omitempty"`
	TaxIncludedPriceMin float64 `json:"tax_included_price_min,omitempty"`
	TaxIncludedPriceMax float64 `json:"tax_included_price_max,omitempty"`
}

type ProductAggregatedInventory struct {
	ProductID         int    `json:"product_id"`
	InventoryLevel    int    `json:"inventory_level"`
	InventoryWarning  int    `json:"inventory_warning_level"`
	WarrantiesCount   int    `json:"warranties_count"`
	VariantsCount     int    `json:"variants_count"`
	InventoryTracking string `json:"inventory_tracking"`
}

type ProductInventoryResponse struct {
	Data ProductAggregatedInventory `json:"data"`
	Meta Meta                       `json:"meta"`
}

type ProductInventoriesResponse struct {
	Data []ProductAggregatedInventory `json:"data"`
	Meta Meta                         `json:"meta"`
}

type RelatedProductsService struct {
	client *Client
}

type ProductChannelAssignmentsService struct {
	client *Client
}

type ProductCategoriesService struct {
	client *Client
}

type BatchService struct {
	client *Client
}

type PricingService struct {
	client *Client
}

type InventoryService struct {
	client *Client
}

func (s *RelatedProductsService) CreateContext(ctx context.Context, productID int, relatedProductIDs []int) (*http.Response, error) {
	path := fmt.Sprintf("catalog/products/%d/related", productID)

	type RelatedProductsRequest struct {
		ProductIDs []int `json:"product_ids"`
	}

	req, err := s.client.NewRequest(ctx, "POST", path, RelatedProductsRequest{ProductIDs: relatedProductIDs})
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

func (s *RelatedProductsService) DeleteContext(ctx context.Context, productID int) (*http.Response, error) {
	path := fmt.Sprintf("catalog/products/%d/related", productID)

	req, err := s.client.NewRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

func (s *RelatedProductsService) DeleteByIDContext(ctx context.Context, productID, relatedProductID int) (*http.Response, error) {
	path := fmt.Sprintf("catalog/products/%d/related/%d", productID, relatedProductID)

	req, err := s.client.NewRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

func (s *ProductChannelAssignmentsService) ListContext(ctx context.Context, productID int) (*ProductChannelAssignmentsResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/channels", productID)

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	assignmentsResponse := new(ProductChannelAssignmentsResponse)
	_, err = s.client.Do(req, assignmentsResponse)
	return assignmentsResponse, err
}

func (s *ProductChannelAssignmentsService) CreateContext(ctx context.Context, productID int, channelIDs []int) (*http.Response, error) {
	path := fmt.Sprintf("catalog/products/%d/channels", productID)

	type ChannelAssignmentsRequest struct {
		ChannelIDs []int `json:"channel_ids"`
	}

	req, err := s.client.NewRequest(ctx, "POST", path, ChannelAssignmentsRequest{ChannelIDs: channelIDs})
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

func (s *ProductChannelAssignmentsService) DeleteChannelContext(ctx context.Context, productID, channelID int) (*http.Response, error) {
	path := fmt.Sprintf("catalog/products/%d/channels/%d", productID, channelID)

	req, err := s.client.NewRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

func (s *ProductCategoriesService) ListContext(ctx context.Context, productID int) (*CategoryAssignmentsResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/categories", productID)

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	assignmentsResponse := new(CategoryAssignmentsResponse)
	_, err = s.client.Do(req, assignmentsResponse)
	return assignmentsResponse, err
}

func (s *ProductCategoriesService) CreateContext(ctx context.Context, productID int, categoryIDs []int) (*http.Response, error) {
	path := fmt.Sprintf("catalog/products/%d/categories", productID)

	type CategoryAssignmentsRequest struct {
		CategoryIDs []int `json:"category_ids"`
	}

	req, err := s.client.NewRequest(ctx, "POST", path, CategoryAssignmentsRequest{CategoryIDs: categoryIDs})
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

func (s *ProductCategoriesService) DeleteContext(ctx context.Context, productID int) (*http.Response, error) {
	path := fmt.Sprintf("catalog/products/%d/categories", productID)

	req, err := s.client.NewRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

func (s *ProductCategoriesService) DeleteCategoryContext(ctx context.Context, productID, categoryID int) (*http.Response, error) {
	path := fmt.Sprintf("catalog/products/%d/categories/%d", productID, categoryID)

	req, err := s.client.NewRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(req, nil)
}

func (s *BatchService) CreateProductsContext(ctx context.Context, products []Product) (*BatchProductsResponse, error) {
	path := "catalog/products"

	req, err := s.client.NewRequest(ctx, "POST", path, BatchProductCreateRequest{Products: products})
	if err != nil {
		return nil, err
	}

	batchResponse := new(BatchProductsResponse)
	_, err = s.client.Do(req, batchResponse)
	return batchResponse, err
}

func (s *BatchService) UpdateProductsContext(ctx context.Context, products []Product) (*BatchProductsResponse, error) {
	path := "catalog/products"

	req, err := s.client.NewRequest(ctx, "PUT", path, BatchProductUpdateRequest{Products: products})
	if err != nil {
		return nil, err
	}

	batchResponse := new(BatchProductsResponse)
	_, err = s.client.Do(req, batchResponse)
	return batchResponse, err
}

func (s *BatchService) DeleteProductsContext(ctx context.Context, productIDs []int) (*BatchErrorResponse, error) {
	path := "catalog/products"

	req, err := s.client.NewRequest(ctx, "DELETE", path, BatchProductDeleteRequest{IDs: productIDs})
	if err != nil {
		return nil, err
	}

	batchResponse := new(BatchErrorResponse)
	_, err = s.client.Do(req, batchResponse)
	return batchResponse, err
}

func (s *PricingService) GetContext(ctx context.Context, request PricingRequest) (*PricingResponse, error) {
	path := "pricing/products"

	req, err := s.client.NewRequest(ctx, "POST", path, request)
	if err != nil {
		return nil, err
	}

	pricingResponse := new(PricingResponse)
	_, err = s.client.Do(req, pricingResponse)
	return pricingResponse, err
}

func (s *InventoryService) GetContext(ctx context.Context, productID int) (*ProductInventoryResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/inventory", productID)

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	inventoryResponse := new(ProductInventoryResponse)
	_, err = s.client.Do(req, inventoryResponse)
	return inventoryResponse, err
}

func (s *InventoryService) ListContext(ctx context.Context, productIDs []int) (*ProductInventoriesResponse, error) {
	path := "catalog/products/inventory"

	values := make(map[string][]string)
	for _, id := range productIDs {
		values["id:in"] = append(values["id:in"], strconv.Itoa(id))
	}

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	for key, vals := range values {
		for _, val := range vals {
			query.Add(key, val)
		}
	}
	req.URL.RawQuery = query.Encode()

	inventoriesResponse := new(ProductInventoriesResponse)
	_, err = s.client.Do(req, inventoriesResponse)
	return inventoriesResponse, err
}

type BulkPricingRulesService struct {
	client *Client
}

func (s *BulkPricingRulesService) ListContext(ctx context.Context, productID int, params *QueryParams) (*PricingRulesResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/bulk-pricing-rules", productID)

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	if params != nil {
		req.URL.RawQuery = params.ToValues().Encode()
	}

	rulesResponse := new(PricingRulesResponse)
	_, err = s.client.Do(req, rulesResponse)
	return rulesResponse, err
}

func (s *BulkPricingRulesService) GetContext(ctx context.Context, productID, ruleID int) (*PricingRuleResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/bulk-pricing-rules/%d", productID, ruleID)

	req, err := s.client.NewRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	ruleResponse := new(PricingRuleResponse)
	_, err = s.client.Do(req, ruleResponse)
	return ruleResponse, err
}

func (s *BulkPricingRulesService) CreateContext(ctx context.Context, productID int, rule *PricingRule) (*PricingRuleResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/bulk-pricing-rules", productID)

	req, err := s.client.NewRequest(ctx, "POST", path, rule)
	if err != nil {
		return nil, err
	}

	ruleResponse := new(PricingRuleResponse)
	_, err = s.client.Do(req, ruleResponse)
	return ruleResponse, err
}

func (s *BulkPricingRulesService) UpdateContext(ctx context.Context, productID, ruleID int, rule *PricingRule) (*PricingRuleResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/bulk-pricing-rules/%d", productID, ruleID)

	req, err := s.client.NewRequest(ctx, "PUT", path, rule)
	if err != nil {
		return nil, err
	}

	ruleResponse := new(PricingRuleResponse)
	_, err = s.client.Do(req, ruleResponse)
	return ruleResponse, err
}

func (s *BulkPricingRulesService) DeleteContext(ctx context.Context, productID, ruleID int) error {
	path := fmt.Sprintf("catalog/products/%d/bulk-pricing-rules/%d", productID, ruleID)

	req, err := s.client.NewRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(req, nil)
	return err
}

func (s *BulkPricingRulesService) UpdateBatchContext(ctx context.Context, productID int, request *BulkPricingRuleRequest) (*BulkPricingRuleResponse, error) {
	path := fmt.Sprintf("catalog/products/%d/bulk-pricing-rules", productID)

	req, err := s.client.NewRequest(ctx, "PUT", path, request)
	if err != nil {
		return nil, err
	}

	bulkResponse := new(BulkPricingRuleResponse)
	_, err = s.client.Do(req, bulkResponse)
	return bulkResponse, err
}

func (s *BulkPricingRulesService) DeleteAllContext(ctx context.Context, productID int) error {
	path := fmt.Sprintf("catalog/products/%d/bulk-pricing-rules", productID)

	req, err := s.client.NewRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	_, err = s.client.Do(req, nil)
	return err
}

type BulkPricingRuleRequest struct {
	BulkPricingRules []PricingRule `json:"bulk_pricing_rules"`
}

type BulkPricingRuleResponse struct {
	Data []PricingRule `json:"data"`
	Meta Meta          `json:"meta"`
}
