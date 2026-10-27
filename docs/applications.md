# Arizalar API

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
    {"field": 20, "name": "backend_language", "value": "Go"},
    {"field": 21, "name": "experience_years", "value": "3"},
    {"field": 22, "name": "technologies", "value": "Docker"},
    {"field": 22, "name": "technologies", "value": "PostgreSQL"},
    {"field": 23, "name": "agree", "value": "true"}
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
yaratadi; `011_add_applications_data_name.sql` unga `name VARCHAR(200) NOT NULL`
ustunini qo'shadi. Natijada jadval `id`, `application_id`, `page`, `slug`, `field`,
`name`, `value` qiymatlarini saqlaydi. `page`
bazada sahifaning IDsi, `slug` shu sahifaning URL orqali yuborilgan slugi,
`field` esa field IDsi, `name` saqlash paytidagi field nomi. Nom bodydan alohida
yuborilmaydi: service uni mavjud formadagi fielddan oladi. Checkboxning har bir
tanlovi uchun ham shu field IDsi va nomi saqlanadi.
`011` migratsiyasi mavjud javoblarning nomini `fields` jadvalidan to'ldiradi;
field allaqachon o'chirilgan bo'lsa, eski javobning nomi bo'sh string qoladi.
Jadvalda `REFERENCES` yoki foreign key yo'q;
ariza, sahifa va field mosligi saqlashdan oldin service orqali tekshiriladi.
Sahifa yoki fieldni keyinchalik o'chirish/ko'chirish bu jadval orqali
bloklanmaydi va saqlangan javoblar avtomatik o'zgarmaydi.

Ariza topilmasa yoki slug shu kategoriyaning faol formasida bo'lmasa `404`;
arizaning turnir/kategoriya IDlari mos kelmasa `400`; kategoriya faol bo'lmasa
`409`; kutilmagan xato `500`.

`010` migratsiya fayli shu jadvalning yangi tuzilishi bilan tahrirlangan.
Oldingi `010` bazada bajarilgan bo'lsa, mavjud migratsiya mexanizmi uni
avtomatik qayta bajarmaydi; ushbu fayl o'zgarishi mavjud bazani o'zi yangilamaydi.

## 4. Arizalar ro'yxati

```http
GET /api/applications?tournament_id=1&category_id=12&email=ali&name=Valiyev&limit=10
Authorization: Bearer <access_token>
```

Ro'yxat va bitta arizani olish autentifikatsiya talab qiladi. Mavjud POST
route'lari ommaviy qoladi.

| Query | Tavsif |
| --- | --- |
| `tournament_id` | Turnir IDsi, musbat butun son; yuborilmasa barcha turnirlar. |
| `category_id` | Kategoriya IDsi, musbat butun son; yuborilmasa barcha kategoriyalar. |
| `email` | Arizaning emaili bo'yicha katta-kichik harfsiz qisman qidirish. |
| `name` | Arizachining ismi bo'yicha katta-kichik harfsiz qisman qidirish. |
| `limit` | `1..100`, standart `10`. |
| `cursor` | Arizaning musbat `id` qiymati. |
| `direction` | `next` (standart) yoki `prev`; `prev` uchun `cursor` majburiy. |

Filtrlar birgalikda ishlaydi. Natija turnir ro'yxatidagi kabi ariza IDsi
kamayish tartibida qaytadi. Keyingi va oldingi sahifalarda filtrlarni saqlang:

```http
GET /api/applications?tournament_id=1&category_id=12&email=ali&name=Valiyev&limit=10&cursor=25&direction=next
GET /api/applications?tournament_id=1&category_id=12&email=ali&name=Valiyev&limit=10&cursor=25&direction=prev
```

Muvaffaqiyatli javob `200`:

```json
{
  "data": [
    {
      "id": 25,
      "tournament_id": 1,
      "tournament_name": "Dasturlash turniri",
      "category_id": 12,
      "category_name": "Backend",
      "name": "Ali Valiyev",
      "email": "ali@example.com",
      "birth_date": "2000-05-18",
      "email_verified_at": "2026-10-22T10:01:00Z",
      "created_at": "2026-10-22T10:00:00Z",
      "updated_at": "2026-10-22T10:01:00Z",
      "data": [
        {
          "slug": "server-texnologiyalari",
          "name": "backend_language",
          "value": "Go"
        }
      ]
    }
  ],
  "next_cursor": null,
  "prev_cursor": null,
  "has_next": false,
  "has_prev": false
}
```

`tournament_name` va `category_name` tegishli turnir va kategoriyadan JOIN
orqali olinadi. GET javoblarida arizaning identifikatori `id` nomi bilan qaytadi.
Har bir arizaning `data` massivi `applications_data` jadvalidan JOIN orqali
olinadi va `page ASC, id ASC` tartibida qaytadi. `name` javob saqlangan paytdagi
field nomidir. Ichki obyektlarda faqat `slug`, `name`, `value` qaytadi.
Bir nechta javobi bor ariza ham ro'yxatda bir marta qaytadi;
`limit` arizalar soniga tegishli. Javoblar saqlanmagan bo'lsa `data: []`.
Ariza topilmagan ro'yxatda tashqi `data: []`, ikkala cursor `null`, ikkala
`has_*` qiymati `false` bo'ladi.

Noto'g'ri query qiymati `400`, token yo'q yoki yaroqsiz bo'lsa `401`, faol
bo'lmagan foydalanuvchi uchun `403`, kutilmagan xato uchun `500` qaytadi.

## 5. Bitta arizani olish

```http
GET /api/applications/25
Authorization: Bearer <access_token>
```

Javob `200`: ro'yxatdagi bitta ariza obyekti bilan bir xil, shu jumladan
`applications_data`dan olingan `data` massivi. Pagination maydonlari bo'lmaydi.
Noto'g'ri ID uchun `400`, ariza mavjud bo'lmasa `404` qaytadi.

GET route'laridagi `applications_data.name` ustuni uchun mavjud
`011_add_applications_data_name.sql` migratsiyasi bajarilgan bo'lishi kerak.

## 6. Arizani barcha sahifalari bilan yangilash

```http
PUT /api/applications/25
Authorization: Bearer <access_token>
Content-Type: application/json
```

```json
{
  "category_id": 12,
  "name": "Ali Valiyev",
  "email": "ali@example.com",
  "birth_date": "2000-05-18",
  "pages": [
    {
      "slug": "server-texnologiyalari",
      "fields": {
        "backend_language": "Go",
        "experience_years": 4,
        "technologies": ["Docker", "PostgreSQL"],
        "agree": true
      }
    },
    {
      "slug": "qoshimcha-malumotlar",
      "fields": {
        "portfolio": "https://example.com/portfolio"
      }
    }
  ]
}
```

Bu API faqat `admin` va `superadmin` uchun. `category_id`, `name`, `email`,
`birth_date` va `pages` majburiy. Turnir IDsi mavjud arizadan olinadi va
o'zgartirilmaydi; bodyda `tournament_id` yuborish `400` qaytaradi.
Ism va email yaratishdagi kabi tozalanadi; tug'ilgan sana kelajakda bo'lmasligi
kerak. Kategoriya faol va arizaning o'sha turniriga tegishli bo'lishi kerak.

`pages` tanlangan kategoriyaning barcha faol sahifalarini aynan bir martadan
qamrab oladi. Sluglar [forma API](page-form.md) javobidan olinadi; massiv
tartibi ixtiyoriy. Misoldagi ikkala sahifa ham tanlangan kategoriyada mavjud
bo'lishi kerak. Sahifada majburiy field bo'lmasa, `fields: {}` yuborish mumkin;
faol sahifalar umuman bo'lmasa, `pages: []` yuboriladi. `pages` yoki `fields`
uchun `null` qabul qilinmaydi.

Har bir sahifaning qiymatlari 3-bo'limdagi mavjud field validatsiyasidan o'tadi.
Yetishmagan, takrorlangan yoki boshqa kategoriyadagi sahifa uchun `400`
qaytadi. Field xatolarida sahifa slugi ham ko'rsatiladi:

```json
{
  "errors": {
    "pages.server-texnologiyalari.backend_language": "field is required",
    "pages.qoshimcha-malumotlar": "page is required"
  }
}
```

Bu to'liq almashtirish amali: barcha eski sahifa javoblari o'chirilib,
yuborilgan qiymatlar yoziladi. Jumladan, eski kategoriyadagi yoki hozir faol
bo'lmagan sahifalarning javoblari ham olib tashlanadi. Yuborilmagan ixtiyoriy
fieldlar va olib tashlangan checkbox tanlovlari saqlanmaydi. Sahifalarni
alohida ketma-ket yuborish talab qilinmaydi. Boshqa arizalar o'zgarmaydi.

Service bitta tranzaksiyada arizaning asosiy ma'lumotlarini yangilaydi va
mavjud `SaveData` repository metodi orqali barcha sahifa javoblarini saqlaydi.
Ariza yozuvi qulflanadi; sahifalardan birini yozishda xato chiqsa, asosiy
ma'lumotlar va barcha javoblar birgalikda bekor qilinadi. Parallel sahifa
saqlash eski kategoriya IDsi bilan kelsa, yangilangan kategoriyaga yozolmaydi.
Forma konfiguratsiyasini parallel o'zgartirishga oid 3-bo'limdagi cheklov
ushbu APIga ham tegishli.

Email o'zgarsa, `email_verified_at` qiymati `null` qilinadi va yangi 6 xonali,
10 daqiqalik tasdiqlash kodi yaratiladi. Hali tasdiqlanmagan arizaning
kategoriyasi o'zgarsa ham, yangi kategoriya uchun kod yangilanadi. Eski kod
almashtiriladi, xato urinishlar soni nolga qaytadi. Kod yaratishdagi kabi
server logiga chiqadi; email yuborilmaydi. Kodni mavjud `verify-email` API
orqali yangi email va o'sha turnir IDsi bilan tasdiqlash mumkin. Email
o'zgarmasa, mavjud tasdiqlangan holat saqlanadi.

Muvaffaqiyatli javob `200`; asosiy ma'lumotlar qaytadi:

```json
{
  "application_id": 25,
  "tournament_id": 1,
  "category_id": 12,
  "name": "Ali Valiyev",
  "email": "ali@example.com",
  "birth_date": "2000-05-18",
  "email_verified_at": "2026-10-22T10:01:00Z",
  "created_at": "2026-10-22T10:00:00Z",
  "updated_at": "2026-10-27T10:00:00Z"
}
```

Yangi kod yaratilganida javobda qo'shimcha `code_expires_at` qaytadi; kodning
o'zi qaytmaydi. Barcha saqlangan javoblarni `GET /api/applications/25` orqali
olish mumkin.

Kod fayliga yozish muvaffaqiyatsiz bo'lsa, baza tranzaksiyasi bekor qilinadi.
Kod yozilib, baza commitida xato chiqsa, service oldingi kod yozuvini tiklashga
urinadi. Baza va fayl bitta atomar tranzaksiya emas; jarayon to'xtashi yoki
tiklash xatosi sababli ular mos kelmay qolishi mumkin. Tiklash xatosi logga
yoziladi. Update va email tasdiqlash mavjud mutex bilan bitta server jarayoni
ichida ketma-ket bajariladi.

| Holat | HTTP status |
| --- | --- |
| Ariza va barcha javoblar yangilandi | `200` |
| Noto'g'ri ID, JSON, sana, sahifa yoki field qiymati | `400` |
| Kategoriya boshqa turnirga tegishli | `400` |
| Token yo'q yoki yaroqsiz | `401` |
| Foydalanuvchi faol emas yoki admin/superadmin emas | `403` |
| Ariza, turnir yoki kategoriya topilmadi | `404` |
| Kategoriya faol emas yoki FK ziddiyati | `409` |
| Bazaga/faylga murojaatda kutilmagan xato | `500` |

Yangi migratsiya talab qilinmaydi; mavjud `008`, `010`, `011` jadvallari
va ustunlari ishlatiladi.
