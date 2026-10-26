# Kategoriyalar ro'yxati

```http
GET /api/categories?tournament_id=1&limit=3
Authorization: Bearer <token>
```

Ro'yxat tournament kabi ikki yo'nalishli cursor pagination bilan ishlaydi.
Natija har ikkala yo'nalishda ham `id DESC` tartibida qaytadi.

| Query | Qiymat |
| --- | --- |
| `direction` | `next` yoki `prev`; standart qiymat `next`. |
| `cursor` | Musbat ID; `prev` uchun majburiy. |
| `limit` | `1` dan `100` gacha; standart qiymat `10`. |
| `tournament_id` | Ixtiyoriy musbat turnir IDsi. |
| `parent_id` | Musbat ID bo'lsa shu kategoriyaning bevosita bolalari; yuborilmasa yoki `0` bo'lsa bosh kategoriyalar. |
| `name` | Nom bo'yicha qidirish. |
| `is_active` | `true` yoki `false`; yuborilmasa ikkala holat ham. |

Masalan, tanlangan filterlarga mos IDlar `100, 99, 98, 97, 96, 95` bo'lsin.
Birinchi so'rov `100, 99, 98`ni va quyidagi pagination maydonlarini qaytaradi:

```json
{
  "next_cursor": 98,
  "prev_cursor": null,
  "has_next": true,
  "has_prev": false
}
```

Yuqorida javobning faqat pagination maydonlari ko'rsatilgan. To'liq javobda
kategoriyalar avvalgidek `data` massivida keladi.

Keyingi sahifa uchun javobdagi `next_cursor` ishlatiladi:

```http
GET /api/categories?tournament_id=1&limit=3&cursor=98&direction=next
Authorization: Bearer <token>
```

Natija `97, 96, 95`; `prev_cursor: 97`, `has_prev: true`,
`next_cursor: null`, `has_next: false` bo'ladi.
Oldingi sahifaga qaytish uchun shu javobdagi `prev_cursor` ishlatiladi:

```http
GET /api/categories?tournament_id=1&limit=3&cursor=97&direction=prev
Authorization: Bearer <token>
```

Natija yana `100, 99, 98` bo'ladi. Sahifalar orasida barcha filterlarni
va `limit`ni saqlang. Tegishli yo'nalishda sahifa bo'lmasa cursor `null`,
`has_next` yoki `has_prev` esa `false` qaytadi. Bo'sh natijada `data: []`,
ikkala cursor `null` va ikkala belgi `false` bo'ladi.

Noto'g'ri query qiymatlari, jumladan `direction=prev` bilan cursor
yuborilmasligi `400` qaytaradi. Ro'yxatni olish autentifikatsiya talab qiladi.
