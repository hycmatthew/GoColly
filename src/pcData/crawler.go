package pcData

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/gocolly/colly/v2"
)

func getUSPriceAndImgFromNewEgg(link string, collector *colly.Collector) (string, string) {
	imgLink, price := "", ""

	collectorErrorHandle(collector, link)

	collector.OnHTML(".is-product", func(element *colly.HTMLElement) {
		imgLink = element.ChildAttr(".swiper-slide .swiper-zoom-container img", "src")
		price = extractFloatStringFromString(element.ChildText(".row-side .product-buy-box li.price-current"))
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

	collector.OnHTML(".product-mallSales", func(element *colly.HTMLElement) {
		price = extractFloatStringFromString(element.ChildText("em.price"))
	})
	if price == "" {
		collector.OnHTML(".product-price-other", func(element *colly.HTMLElement) {
			price = extractFloatStringFromString(element.ChildText("span"))
		})
	}

	collector.Visit(link)
	return price
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