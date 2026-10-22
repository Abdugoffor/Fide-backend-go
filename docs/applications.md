# Ariza yaratish va emailni tasdiqlash

## 1. Ariza yaratish

```http
POST /api/applications
Content-Type: application/json
```

```json
{
  "tournament_id": 1,
  "category_id": 12,
  "name": "Ali Valiyev",
  "email": "ali@example.com",
  "birth_date": "2000-05-18"
}
```

`birth_date` formati `YYYY-MM-DD`; kelajak sana qabul qilinmaydi.
Kategoriya mavjud, faol va shu turnirga tegishli bo'lishi kerak.
Ismdagi tashqi bo'shliqlar olib tashlanadi, email kichik harfga o'tkaziladi.

Muvaffaqiyatli javob — `201`:

```json
{
  "application_id": 25,
  "tournament_id": 1,
  "category_id": 12,
  "name": "Ali Valiyev",
  "email": "ali@example.com",
  "birth_date": "2000-05-18",
  "email_verified_at": null,
  "created_at": "2026-10-22T10:00:00Z",
  "updated_at": "2026-10-22T10:00:00Z",
  "code_expires_at": "2026-10-22T10:10:00Z"
}
```

Server kriptografik tasodifiy 6 xonali kod yaratadi. Kod 10 daqiqa amal qiladi.
Kod HTTP javobiga qo'shilmaydi; hozircha email yuborilmaydi.
Server logida quyidagi ko'rinishda chiqadi:

```text
Application verification: tournament_id=1 category_id=12 application_id=25 code=012345
```

Barcha kutilayotgan kodlar bitta faylda saqlanadi:

```text
.local/application_codes.json
```

Manzilni `.env` orqali o'zgartirish mumkin:

```dotenv
APPLICATION_CODES_FILE=.local/application_codes.json
```

Fayl birinchi arizada avtomatik yaratiladi. Unda JSON massiv saqlanadi:

```json
[
  {
    "tournament_id": 1,
    "category_id": 12,
    "application_id": 25,
    "code": "012345",
    "expires_at": "2026-10-22T10:10:00Z",
    "attempts": 0
  }
]
```

Fayl server qayta ishga tushganda ham saqlanadi; shu bazaga shu fayl bilan
qayta ulanish kerak. Faylga yozish bitta jarayon ichida mutex bilan himoyalangan
va vaqtinchalik faylni almashtirish orqali bajariladi. Bu saqlash usuli bitta
server jarayoni uchun mo'ljallangan; `.lock` fayllari ishlatilmaydi.

## 2. Ikkinchi sahifada kodni tekshirish

```http
POST /api/applications/verify-email
Content-Type: application/json
```

```json
{
  "tournament_id": 1,
  "email": "ali@example.com",
  "application_id": 25,
  "code": "012345"
}
```

`code` string bo'lishi kerak: boshidagi `0` saqlanishi uchun uni son qilib yubormang.
`category_id` yuborilmaydi; service uni arizaning bazadagi yozuvidan oladi.
Email va turnir IDsi arizaga mos bo'lishi, fayldagi turnir/kategoriya ham shu
arizaga tegishli bo'lishi kerak.

Kod to'g'ri bo'lsa, arizaning `email_verified_at` maydoni belgilanadi, kod
fayldan o'chiriladi va `200` bilan [mavjud forma javobi](page-form.md) qaytadi:

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
        }
      ]
    }
  ]
}
```

Sahifa va maydonlar `sort ASC` bo'yicha tartiblanadi. Javobga `sort`, `kind`,
`tournament_id`, `category_id` qo'shilmaydi. Faol sahifalar bo'lmasa,
`{"total":0,"data":[]}`; faol maydonsiz sahifada `fields: []`.

Kod bir marta ishlaydi. Bir ariza uchun parallel tasdiqlashlardan faqat bittasi
muvaffaqiyatli bo'ladi; keyingi so'rovlar `409` qaytaradi. 5 ta noto'g'ri koddan
keyin shu arizani tasdiqlash bloklanadi. Kodni qayta yuborish API'si hozircha yo'q.

Ikkala API ommaviy: admin login yoki Bearer token talab qilmaydi. Tasdiqlashda
email va bir martalik kod tekshiriladi. Mavjud `/api/pages-form`ning auth qoidalari
saqlanadi.

| Holat | HTTP status |
| --- | --- |
| Ariza yaratildi | `201` |
| Kod tasdiqlandi, forma qaytdi | `200` |
| Noto'g'ri JSON, ID, email yoki tug'ilgan sana | `400` |
| Kod xato, yo'q, muddati o'tgan yoki arizaga mos emas | `400` |
| Kategoriya boshqa turnirga tegishli | `400` |
| Ariza, turnir yoki kategoriya topilmadi | `404` |
| Ariza oldin tasdiqlangan yoki kategoriya faol emas | `409` |
| 5 ta noto'g'ri kod | `429` |
| Bazaga/faylga murojaatda kutilmagan xato | `500` |

Arizalar `008_create_applications.sql` migratsiyasidagi `applications` jadvalida
saqlanadi. Migratsiya mavjud `config.RunMigrations` orqali ilova ishga tushganda
bajariladi. Arizalari bor kategoriya yoki turnirni o'chirish, kategoriyani boshqa
turnirga ko'chirish `409`; bog'liq arizalar avtomatik o'chirilmaydi.

Ariza bazaga yozilgandan keyin kod faylini saqlash muvaffaqiyatsiz bo'lsa,
service yangi arizani o'chirishga urinadi va `500` qaytaradi. Baza va fayl bitta
tranzaksiyaga kirmaydi; tozalash xatosi logga yoziladi. Tasdiqlangan arizaning
kodi fayldan o'chmay qolsa ham, bazadagi tasdiqlash belgisi uni qayta ishlatishga
yo'l qo'ymaydi.
