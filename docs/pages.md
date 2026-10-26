# Page API va avtomatik slug

Sahifa yaratilganda `slug` serverda `name`dan avtomatik hosil qilinadi.
Slug faqat kichik lotin harflari va `-`dan iborat: masalan,
`Shaxsiy ma'lumotlar` nomidan `shaxsiy-malumotlar` hosil bo'ladi.
Bo'sh joylar va boshqa belgilar `-` bilan ajratiladi; apostroflar olib tashlanadi.
Lotin harfi bo'lmagan nom uchun asosiy qiymat `page` bo'ladi.

Slug barcha sahifalar bo'yicha unique. Nom takrorlansa yoki hosil bo'lgan slug
band bo'lsa, `-a`, `-b`, ..., `-z`, `-aa` kabi harfli qo'shimcha qo'shiladi.
Parallel yaratishda ham takrorlanishni baza cheklovi to'xtatadi.

Create va update bodylarida `slug` yuborilmaydi. Noma'lum maydon sifatida
yuborilgan `slug` `400` qaytaradi. Sahifa nomi yoki boshqa ma'lumotlari
yangilanganda mavjud slug saqlanadi.

```http
POST /api/pages
Authorization: Bearer <admin-token>
Content-Type: application/json
```

```json
{
  "tournament_id": 1,
  "category_id": 12,
  "name": "Shaxsiy ma'lumotlar",
  "sort": 3,
  "is_active": true
}
```

Create `201` bilan sahifani qaytaradi:

```json
{
  "id": 10,
  "tournament_id": 1,
  "category_id": 12,
  "name": "Shaxsiy ma'lumotlar",
  "slug": "shaxsiy-malumotlar",
  "sort": 3,
  "is_active": true,
  "created_at": "2026-10-26T08:00:00Z",
  "updated_at": "2026-10-26T08:00:00Z"
}
```

`PUT /api/pages/:id` shu body maydonlarini qabul qiladi va `200` bilan sahifani
qaytaradi. `GET /api/pages/:id` ham shu javob shaklida ishlaydi.
`GET /api/pages?tournament_id=1&category_id=12` javobidagi `data` massivining
har bir sahifasida ham `slug` bor. List filterlari va tartibi saqlanadi.

`GET /api/pages-form` va ariza emailini tasdiqlashdan qaytadigan
[forma javobida](page-form.md) ham har bir sahifaning `slug` maydoni mavjud.

`009_add_page_slug.sql` migratsiyasi mavjud sahifalar uchun nomdan slug
hosil qilib, takrorlanmasligi uchun IDning harfli ko'rinishini qo'shadi.
Masalan, ID `10` bo'lsa qo'shimcha `-ba` bo'ladi. Migratsiya mavjud mexanizm
orqali server ishga tushishida qo'llanadi.
