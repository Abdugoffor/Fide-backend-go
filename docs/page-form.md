# Kategoriya sahifalari va maydonlari

```http
GET /api/pages-form?tournament_id=1&category_id=12
Authorization: Bearer <token>
```

`tournament_id` va `category_id` queryda majburiy musbat IDlar. Kategoriya
tanlangan turnirga tegishli va faol bo'lishi kerak.

API shu turnir va kategoriyaga tegishli bazadagi faol sahifalarni, har bir sahifa
ichida esa uning faol maydonlarini qaytaradi. Sahifalar ham, maydonlar ham SQLda
`sort ASC, id ASC` bo'yicha tartiblanadi. Javobda `tournament_id`, `category_id`
va `sort` ko'rsatilmaydi. `total` qaytgan sahifalar sonini bildiradi.
Sahifa nomi `page_name`da qaytadi. Har bir field maydonlari `id`, `type`, `name`,
`validation`, `options` ketma-ketligida yoziladi.

```json
{
  "total": 1,
  "data": [
    {
      "id": 10,
      "page_name": "Server texnologiyalari",
      "fields": [
        {
          "id": 20,
          "type": "select",
          "name": "backend_language",
          "validation": {"required": true},
          "options": ["Go", "Python"]
        },
        {
          "id": 21,
          "type": "number",
          "name": "experience_years",
          "validation": {"required": true, "min": 0, "max": 50},
          "options": []
        }
      ]
    }
  ]
}
```

Faol sahifa bo'lmasa, `{"total":0,"data":[]}`. Sahifada faol maydon bo'lmasa,
`fields: []`. Ota kategoriya sahifalari meros olinmaydi, qo'shimcha doimiy
bosqichlar qo'shilmaydi. CRUD APIlaridagi `sort` maydonlari saqlanadi.

ID noto'g'ri yoki yuborilmagan bo'lsa — `400`; turnir/kategoriya topilmasa — `404`;
kategoriya boshqa turnirga tegishli bo'lsa — `400`; faol bo'lmagan kategoriya — `409`.
