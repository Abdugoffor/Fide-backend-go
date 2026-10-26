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
      "slug": "server-texnologiyalari",
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

## 3. Sahifa javoblarini yaratish yoki yangilash

```http
POST /api/applications/data/server-texnologiyalari
Content-Type: application/json
```

```json
{
  "application_id": 25,
  "tournament_id": 1,
  "category_id": 12,
  "fields": {
    "backend_language": "Go",
    "experience_years": 3,
    "technologies": ["Docker", "PostgreSQL"],
    "agree": true
  }
}
```

Route: `POST /api/applications/data/:slug`. `slug` URLda forma javobidagi
sahifaning `slug` qiymati bilan yuboriladi; bodyda `page` yoki `slug` yuborilmaydi.
`fields` kalitlari shu
sahifaning faol fieldlaridagi `name` qiymatlariga teng bo'lishi kerak. Misoldagi
`technologies` — `options: ["Docker", "PostgreSQL"]` berilgan checkbox,
`agree` esa options berilmagan checkbox. Bu fieldlar bazada mavjud bo'lishi kerak.

Avval `applications` jadvalidagi ariza topilib, uning `tournament_id` va
`category_id` qiymatlari so'rov bilan solishtiriladi. Keyin mavjud
`PageService.Form` orqali kategoriya faolligi, turnirga tegishliligi va
sahifaning shu kategoriyadagi faol forma tarkibida borligi tekshiriladi.
Bu API ham ommaviy; Bearer token, email yoki tasdiqlash kodi so'ramaydi va
`email_verified_at` bo'yicha qo'shimcha cheklov qo'ymaydi.

JSON, DTO teglari va route slugi handlerda bir marta tekshiriladi. Service
bu tekshiruvlarni takrorlamaydi; arizaning tegishliligi, faol forma va
field qiymatlarining biznes validatsiyasini bajaradi. Application, category,
page va field modullaridagi takroriy DTO validatsiyalari ham shu tartibga keltirilgan.

Qiymatlar uchun qoidalar:

- `required` field tushirib qoldirilmaydi; bo'sh string, bo'sh checkbox massivi
  va optionssiz checkbox uchun `false` majburiy qiymat hisoblanmaydi.
- `text`, `textarea`, `tel`, `email`, `date`, `select`, `radio`, `file` string
  qabul qiladi. Stringning chetidagi bo'shliqlar olib tashlanadi.
- `number` JSON son yoki son yozilgan string qabul qiladi; `min` va `max`
  chegaralari tekshiriladi. `NaN` va cheksiz qiymatlar qabul qilinmaydi.
- `min_length` va `max_length` belgilar soni bo'yicha tekshiriladi.
- `email` email formatiga, `date` haqiqiy `YYYY-MM-DD` sanaga mos bo'ladi.
  `tel` uchun matn uzunligi qoidalari ishlaydi; alohida telefon formati belgilanmagan.
- `select` va `radio` qiymati `options` ichida bo'ladi.
- Optionsli `checkbox` stringlar massivini qabul qiladi; har bir qiymat
  `options` ichida va takrorlanmagan bo'ladi. Har bir tanlov alohida qatorga yoziladi.
- Optionssiz `checkbox` boolean qabul qiladi: `true` uchun `"true"` qatori
  saqlanadi, `false` uchun qator yaratilmaydi.
- `file` hozircha fayl manzili yoki identifikatori sifatida string saqlaydi;
  bu endpoint fayl yuklamaydi yoki fayl mavjudligini tekshirmaydi.
- Begona yoki faol bo'lmagan field, noto'g'ri qiymat turi va `null` qabul
  qilinmaydi. Ixtiyoriy fieldni tushirib qoldirish yoki bo'sh yuborish mumkin.
  `fields` obyektining o'zi majburiy; `{}` faqat majburiy fieldlar bo'lmasa o'tadi.

Validatsiya xatolari `400` bilan qaytadi, bazadagi oldingi javoblar saqlanadi:

```json
{
  "errors": {
    "experience_years": "must be at most 50",
    "backend_language": "field is required"
  }
}
```

Muvaffaqiyatli javob — shu ariza va sahifa uchun avval qatorlar bo'lmasa `201`,
avvalgi qatorlar almashtirilsa `200`:

```json
{
  "application_id": 25,
  "page": "server-texnologiyalari",
  "data": [
    {"field": 20, "value": "Go"},
    {"field": 21, "value": "3"},
    {"field": 22, "value": "Docker"},
    {"field": 22, "value": "PostgreSQL"},
    {"field": 23, "value": "true"}
  ]
}
```

Har so'rov bitta sahifaning barcha javoblarini almashtiradi. Yuborilmagan
ixtiyoriy fieldlar va olib tashlangan checkbox tanlovlari o'chadi. Boshqa
sahifalar va boshqa arizalarning javoblari o'zgarmaydi. Barcha qiymatlar bo'sh
bo'lsa qator qolmaydi; shundan keyingi bo'sh saqlash ham `201` qaytaradi.

Service tranzaksiyani boshlaydi va yakunlaydi; repository shu tranzaksiyada
arizani `FOR UPDATE` bilan qulflaydi, shu sahifaning oldingi qatorlarini o'chiradi
va yangi qiymatlarni yozadi. Xatoda tranzaksiya bekor qilinadi; yarim yozilgan
javob qolmaydi. Bir arizaga parallel yozishlar ketma-ket bajariladi, oxirgi
saqlangan sahifa javobi qoladi. Forma konfiguratsiyasi validatsiyadan oldin
o'qiladi; administratorning parallel konfiguratsiya o'zgarishlari bilan
validatsiya va saqlash bitta snapshot bo'lishi kafolatlanmaydi.

`010_create_applications_data.sql` migratsiyasi `applications_data` jadvalini
yaratadi: `id`, `application_id`, `page`, `slug`, `field`, `value`. `page`
bazada sahifaning IDsi, `slug` shu sahifaning URL orqali yuborilgan slugi,
`field` esa field IDsi. Jadvalda `REFERENCES` yoki foreign key yo'q;
ariza, sahifa va field mosligi saqlashdan oldin service orqali tekshiriladi.
Sahifa yoki fieldni keyinchalik o'chirish/ko'chirish bu jadval orqali
bloklanmaydi va saqlangan javoblar avtomatik o'zgarmaydi.

Ariza topilmasa yoki slug shu kategoriyaning faol formasida bo'lmasa `404`;
arizaning turnir/kategoriya IDlari mos kelmasa `400`; kategoriya faol bo'lmasa
`409`; kutilmagan xato `500`.

`010` migratsiya fayli shu jadvalning yangi tuzilishi bilan tahrirlangan.
Oldingi `010` bazada bajarilgan bo'lsa, mavjud migratsiya mexanizmi uni
avtomatik qayta bajarmaydi; ushbu fayl o'zgarishi mavjud bazani o'zi yangilamaydi.
