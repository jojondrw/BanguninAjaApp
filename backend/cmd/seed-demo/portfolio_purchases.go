package main

import (
	"fmt"
	"math"
)

var procurementMaterials = []struct {
	code  string
	price int64
}{
	{materialCement, 74_000}, {materialRebar13, 167_000}, {materialReadymix, 1_160_000}, {materialLightBrick, 655_000},
	{materialTile, 69_000}, {materialCable, 9_700}, {materialWF, 4_820_000}, {materialPaving, 97_000},
}

type generatedPurchase struct {
	request requestSeed
	order   orderSeed
	receipt *receiptSeed
	payable payableSeed
}

func generatedPurchases() []generatedPurchase {
	vendors := []string{vendorBuilding, vendorSteel, vendorConcrete, vendorSurabaya, vendorMakassar, vendorGeneral}
	purchases := []generatedPurchase{}
	for index, seed := range generatedSeeds() {
		if seed.status != "ongoing" {
			continue
		}
		code := generatedCode(seed)
		for batch := range 3 {
			material := procurementMaterials[(index+batch*3)%len(procurementMaterials)]
			quantity := float64(80 + (index*37+batch*53)%400)
			day := -(12 + batch*45 + index%9)
			number := fmt.Sprintf("G%s-%d", seed.suffix, batch+1)
			request := requestSeed{"DEMO-PR-" + number, code, day - 4, "completed", fmt.Sprintf("Kebutuhan material batch %d %s", batch+1, seed.name), []lineSeed{{material.code, quantity, 0}}}
			order := orderSeed{"DEMO-PO-" + number, vendors[(index+batch)%len(vendors)], code, request.number, day, 30, true, []lineSeed{{material.code, quantity, material.price}}}
			total := int64(quantity) * material.price
			var receipt *receiptSeed
			switch (index + batch) % 4 {
			case 0, 1:
				receipt = &receiptSeed{"DEMO-GR-" + number, order.number, warehouseCentral, day + 6, "good", "Diterima lengkap", []receiptLineSeed{{material.code, quantity, 0}}}
			case 2:
				rejected := math.Floor(quantity / 20)
				receipt = &receiptSeed{"DEMO-GR-" + number, order.number, warehouseCentral, day + 8, "partially_damaged", "Sebagian rusak saat pengiriman", []receiptLineSeed{{material.code, quantity - rejected, rejected}}}
			}
			due := day + 30
			var payments []paymentSeed
			switch (index + batch) % 5 {
			case 0, 1:
				payments = []paymentSeed{{total, min(due-2, -1), "transfer", "TRF-" + number, "Pelunasan " + order.number}}
			case 2:
				payments = []paymentSeed{{total / 2, min(due-5, -1), "cek", "GIRO-" + number, "Pembayaran sebagian " + order.number}}
			}
			payable := payableSeed{"DEMO-HUT-" + number, order.vendor, code, order.number, due, total, payments}
			purchases = append(purchases, generatedPurchase{request, order, receipt, payable})
		}
	}
	return purchases
}

func generatedRequests() []requestSeed {
	requests := []requestSeed{}
	for _, purchase := range generatedPurchases() {
		requests = append(requests, purchase.request)
	}
	return requests
}

func generatedOrders() []orderSeed {
	orders := []orderSeed{}
	for _, purchase := range generatedPurchases() {
		orders = append(orders, purchase.order)
	}
	return orders
}

func generatedReceipts() []receiptSeed {
	receipts := []receiptSeed{}
	for _, purchase := range generatedPurchases() {
		if purchase.receipt != nil {
			receipts = append(receipts, *purchase.receipt)
		}
	}
	return receipts
}

func generatedPayables() []payableSeed {
	payables := []payableSeed{}
	for _, purchase := range generatedPurchases() {
		payables = append(payables, purchase.payable)
	}
	return payables
}
