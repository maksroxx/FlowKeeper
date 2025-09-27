package stock_test

// func TestLargeDocument_Integration(t *testing.T) {
// 	router, db := setupTestRouter("large_doc_test_db")
// 	sqlDB, _ := db.DB()
// 	defer sqlDB.Close()

// 	h := NewTestHelper(t, router)

// 	const itemCount = 200

// 	t.Log("Подготовка мира: создание склада, справочников и товаров...")
// 	warehouse := h.CreateWarehouse("Центральный распределительный склад")
// 	unit := h.CreateUnit("шт")
// 	category := h.CreateCategory("Разные товары")

// 	items := make([]models.Item, itemCount)
// 	docItems := make([]models.DocumentItem, itemCount)

// 	expectedQuantity := decimal.NewFromInt(10)
// 	expectedPrice := decimal.NewFromFloat(15.5)
// 	expectedTotalCost := expectedQuantity.Mul(expectedPrice)

// 	for i := 0; i < itemCount; i++ {
// 		itemName := fmt.Sprintf("Тестовый товар %d", i+1)
// 		itemSKU := fmt.Sprintf("TEST-SKU-%04d", i+1)

// 		items[i] = h.CreateItem(gin.H{
// 			"name":        itemName,
// 			"sku":         itemSKU,
// 			"unit_id":     unit.ID,
// 			"category_id": category.ID,
// 		})

// 		docItems[i] = models.DocumentItem{
// 			ItemID:   items[i].ID,
// 			Quantity: expectedQuantity,
// 			Price:    decimalPtr(expectedPrice),
// 		}
// 	}
// 	t.Logf("Создано %d товаров.", itemCount)

// 	t.Logf("Создание документа 'Приход' с %d позициями...", itemCount)
// 	incomeDocPayload := models.Document{
// 		Type:        "INCOME",
// 		WarehouseID: &warehouse.ID,
// 		Items:       docItems,
// 	}
// 	incomeDoc := h.CreateDocument(incomeDocPayload)
// 	h.PostDocument(incomeDoc.ID)
// 	t.Log("Проведение завершено.")

// 	t.Log("Проверка остатков после проведения...")
// 	balancesAfterPost := h.GetBalances(warehouse.ID)
// 	h.Assert.Len(balancesAfterPost, itemCount, "Количество записей об остатках должно совпадать с количеством товаров")

// 	balanceMap := make(map[uint]models.StockBalance)
// 	for _, b := range balancesAfterPost {
// 		balanceMap[b.ItemID] = b
// 	}

// 	for _, item := range items {
// 		balance, ok := balanceMap[item.ID]
// 		h.Assert.True(ok, "Не найдена запись об остатке для товара ID %d", item.ID)
// 		h.Assert.True(expectedQuantity.Equal(balance.Quantity), "Неверный остаток для товара ID %d", item.ID)
// 		h.Assert.True(expectedTotalCost.Equal(balance.TotalCost), "Неверная себестоимость для товара ID %d", item.ID)
// 	}
// 	t.Log("Все остатки и себестоимость после проведения корректны.")

// 	t.Log("Отмена проведения большого документа...")
// 	h.CancelDocument(incomeDoc.ID)
// 	t.Log("Отмена завершена.")

// 	t.Log("Проверка остатков после отмены...")
// 	balancesAfterCancel := h.GetBalances(warehouse.ID)
// 	h.Assert.Len(balancesAfterCancel, itemCount)

// 	balanceMapAfterCancel := make(map[uint]models.StockBalance)
// 	for _, b := range balancesAfterCancel {
// 		balanceMapAfterCancel[b.ItemID] = b
// 	}

// 	for _, item := range items {
// 		balance, ok := balanceMapAfterCancel[item.ID]
// 		h.Assert.True(ok, "Не найдена запись об остатке для товара ID %d после отмены", item.ID)
// 		h.Assert.True(decimal.Zero.Equal(balance.Quantity), "Остаток для товара ID %d должен был вернуться к 0", item.ID)
// 		h.Assert.True(decimal.Zero.Equal(balance.TotalCost), "Себестоимость для товара ID %d должна была вернуться к 0", item.ID)
// 	}
// 	t.Log("Все остатки и себестоимость после отмены корректны.")
// }
