package pcData

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly/v2"
	"github.com/imroc/req/v3"
)

func getUSPriceAndImgFromNewEgg(link string, collector *colly.Collector) (string, string) {
	imgLink, price, updatedUrl := "", "", link

	collectorErrorHandle(collector, link)
	collector.OnError(func(r *colly.Response, err error) {
		if r != nil && r.StatusCode == 404 {
			fmt.Println("404 error:", r.Request.URL)
			imgLink = "404"
			price = ""
		}
	})
	collector.OnResponse(func(r *colly.Response) {
		updatedUrl = r.Request.URL.String()
		fmt.Println("Response received:", updatedUrl)
	})

	collector.OnHTML(".is-product", func(element *colly.HTMLElement) {
		imgLink = element.ChildAttr(".swiper-slide .swiper-zoom-container img", "src")
		price = extractFloatStringFromString(element.ChildText(".row-side .product-buy-box .price-current"))
		available := element.ChildText(".row-side .product-buy-box .product-buy .btn-message")
		price = OutOfStockLogic(price, available)
	})

	collector.Visit(link)
	return price, imgLink
}

func getHKPrice(link string, collector *colly.Collector) string {
	price := ""
	collectorErrorHandle(collector, link)

	collector.OnHTML(".line-05", func(element *colly.HTMLElement) {

		element.ForEach(".product-price", func(i int, item *colly.HTMLElement) {
			if price == "" {
				price = extractFloatStringFromString(element.ChildText("span"))
			}
		})
	})

	collector.Visit(link)
	return price
}

func getCNNameAndPriceFromPcOnline(link string, collector *colly.Collector) (string, string) {
	price := ""
	name := ""

	collectorErrorHandle(collector, link)

	collector.OnHTML(".pro-info", func(element *colly.HTMLElement) {
		name = convertGBKString(element.ChildText(".pro-tit h1"))
	})

	collector.OnHTML(".product-price, .price-info", func(element *colly.HTMLElement) {
		mallPrice := extractFloatStringFromString(element.ChildText(".product-mallSales em.price"))

		otherPrice := extractFloatStringFromString(element.ChildText(".product-price-other span"))

		normalPrice := extractFloatStringFromString(element.ChildText(".r-price a"))

		if mallPrice != "" {
			price = mallPrice
		} else if otherPrice != "" {
			price = otherPrice
		} else {
			price = normalPrice
		}
	})
	collector.Visit(link)
	fmt.Println(name, ":", price)
	return name, price
}

func getDetailsLinkFromZol(link string, collector *colly.Collector) string {
	cnLink := "https://detail.zol.com.cn"

	collectorErrorHandle(collector, link)
	collector.OnHTML(".wrapper", func(element *colly.HTMLElement) {
		cnLink += element.ChildAttr(".section-header-link .more", "href")
	})
	collector.Visit(link)
	return cnLink
}

func extractJDPriceFromZol(element *colly.HTMLElement) PriceType {
	// 封裝選擇器
	selectors := struct {
		MallPrice    string
		NormalPrice  string
		LinkSelector string
	}{
		".side .goods-card .item-b2cprice span",
		".side .goods-card .goods-card__price span",
		".side .goods-card .item-b2cprice a",
	}

	// 提取價格
	extractPrice := func(selector string) string {
		price := extractFloatStringFromString(element.ChildText(selector))
		return price
	}

	mallPrice := extractPrice(selectors.MallPrice)
	normalPrice := extractPrice(selectors.NormalPrice)

	// 鏈接提取
	jdLink := GetJDPriceLinkFromZol(
		element.ChildAttr(selectors.LinkSelector, "href"),
	)

	return PriceType{
		Region:    "CN",
		Platform:  Platform_JD,
		Price:     firstNonEmpty(mallPrice, normalPrice, ""),
		PriceLink: jdLink,
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func CreateCollector() *colly.Collector {
	fakeChrome := req.DefaultClient().ImpersonateChrome()

	collector := colly.NewCollector(
		colly.UserAgent(fakeChrome.Headers.Get("user-agent")),
		colly.AllowedDomains(
			"nanoreview.net",
			"www.newegg.com",
			"newegg.com",
			"www.price.com.hk",
			"price.com.hk",
			"www.colorful.cn",
			"colorful.cn",
			"detail.zol.com.cn",
			"zol.com.cn",
			"product.pconline.com.cn",
			"pconline.com.cn",
			"pangoly.com",
			// gpu
			"www.techpowerup.com",
			"techpowerup.com",
			// motherboard
			"asus.com",
			"www.asus.com",
			"tw.msi.com",
			"www.msi.com",
			"msi.com",
		),
		colly.AllowURLRevisit(),
	)
	collector.SetClient(&http.Client{
		Transport: fakeChrome.Transport,
	})
	return collector
}

func collectorErrorHandle(collector *colly.Collector, link string) {
	collector.OnRequest(func(r *colly.Request) {
		// USER_AGENT = 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_10_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/55.0.2883.95 Safari/537.36'
		r.Headers.Set("Connection", "keep-alive")
		r.Headers.Set("Accept", "*/*")
	})

	collector.OnError(func(response *colly.Response, err error) {
		fmt.Println("请求期间发生错误,则调用:", err, " - link: ", link)
	})

	collector.OnResponse(func(response *colly.Response) {
		fmt.Println("收到响应后调用:", response.Request.URL)
	})
}

func GetRamCNPriceFromChromedp(link string) string {
	fmt.Println(link)
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// create chrome instance
	ctx, cancel := chromedp.NewContext(
		allocCtx,
	)
	defer cancel()

	// create a timeout
	ctx, cancel = context.WithTimeout(ctx, 600*time.Second)
	defer cancel()

	// navigate to a page, wait for an element, click
	var cnPrice string
	err := chromedp.Run(ctx,
		chromedp.Navigate(link),
		// wait for footer element is visible (ie, page is loaded)
		chromedp.Sleep(600*time.Second),
		// retrieve the value of the textarea
		chromedp.Value(`.text--fZ9NUhyQ`, &cnPrice),
	)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(cnPrice)
	return cnPrice
}
