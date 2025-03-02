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
	imgLink, price := "", ""

	collectorErrorHandle(collector, link)

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

func getCNPriceFromPcOnline(link string, collector *colly.Collector) string {
	price := ""

	collectorErrorHandle(collector, link)

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
		fmt.Println(price)
	})

	collector.Visit(link)
	return price
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

func getCNPriceFromZol(link string, collector *colly.Collector) string {
	price := ""

	collectorErrorHandle(collector, link)

	collector.OnHTML(".wrapper", func(element *colly.HTMLElement) {
		mallPrice := extractFloatStringFromString(element.ChildText("side .goods-card .item-b2cprice span"))
		// otherPrice := extractFloatStringFromString(element.ChildText(".price__merchant .price"))
		normalPrice := extractFloatStringFromString(element.ChildText(".side .goods-card .goods-card__price span"))

		if mallPrice != "" {
			price = mallPrice
		} else {
			price = normalPrice
		}
	})

	collector.Visit(link)
	return price
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
