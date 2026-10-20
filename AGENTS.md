# Loyiha bo'yicha ishlash qoidalari

Bu fayl butun loyiha uchun qo'llanadi. Yangi kod yozishdan oldin shu qoidalarni va o'zgartiriladigan modulni o'qing. Arxitektura va kod uslubi uchun asosiy namuna — `module/tournament` (API yo'li: `/api/tournaments`).

## Asosiy talablar

- Barcha funksionallikni `handler -> service -> repository` qatlamlariga ajrating. Bu talab faqat CRUD uchun emas: login, logout, qidirish, holat almashtirish, tasdiqlash, ro'yxatdan o'tish va boshqa biznes amallariga ham tegishli.
- Handler biznes qoidalarini bajarmaydi va repository yoki bazaga to'g'ridan-to'g'ri murojaat qilmaydi.
- Service biznes jarayonini boshqaradi; ma'lumotlarni repository orqali oladi va saqlaydi.
- Repository SQL va ma'lumot saqlash bilan ishlaydi; HTTP javoblari yoki biznes jarayonini boshqarmaydi.
- Yangi funksiya, SQL yoki yordamchi yozishdan oldin mavjud funksiyalarni qidiring. Vazifaga mos funksiya bo'lsa, undan foydalaning.
- Zarur bo'lsa boshqa modulning mavjud service yoki repository imkoniyatlarini qayta ishlating. Qatlam chegaralarini va biznes tekshiruvlarini saqlang.
- `tournament`dagi nomlash, qavslar, mantiqiy bloklar va xatolarni tekshirish uslubini davom ettiring.
- Kodni oddiy va tushunarli yozing. Query filterlarini `tournament.List`dagi kabi bevosita `List` handlerida har bir maydon uchun ketma-ket tekshiring. Ular uchun alohida `parseFilter` yoki struct/pointerlar orqali umumiy parse sikli yaratmang.
- Repositoryda `Scan`ni turnirdagi kabi har bir metod ichida ochiq yozing; alohida scan funksiyasi yaratmang. SQL chaqiruvidan keyin oddiy `if err != nil` yetarli, uni qo'shimcha `{ ... }` blokka o'rash shart emas.
- Bu loyihada test fayllarini qoldirmang. Odatda yangi test fayli yozmasdan ishlashini tekshiring; vaqtinchalik test fayli zarur bo'lsa, ishlatib bo'lgach o'chiring.
- Loyihada `.lock` yoki `*.lock` fayllarini yaratmang va qoldirmang.
- Alohida API so'ralsa, uni tegishli modulning route blokida o'z handleriga to'g'ridan-to'g'ri ulang. Boshqa amalning handleri ichida path nomini tekshirib yo'naltirmang. Router cheklovida mavjud CRUD yo'llarini saqlab, to'qnashmaydigan alohida API manzilini ishlating va manzilni aniq ayting.

## Loyihaning hozirgi tuzilishi

Go modul nomi `auth-login`; `go.mod`dagi Go versiyasi `1.24.4`. HTTP uchun `net/http` va `httprouter`, PostgreSQL uchun `pgx/v5` va `pgxpool`, validatsiya uchun `validator/v10` ishlatiladi.

```text
main.go                         # Ilovani ishga tushirish va modullarni ulash
config/                         # Sozlamalar, DB, migratsiyalar, seeder kirish nuqtasi
  migrations/                   # Raqamlangan SQL migratsiyalar
helper/                         # Umumiy JSON, validatsiya, env, parol va token funksiyalari
module/
  auth/
    cmd.go
    dto/req_dto.go
    dto/res_dto.go
    handler/auth_handler.go
    service/auth_service.go
    repository/auth_repository.go
  tournament/
    cmd.go
    dto/req_dto.go
    dto/res_dto.go
    handler/tournament_handler.go
    service/tournament_service.go
    repository/tournament_repository.go
  category/
    cmd.go
    dto/req_dto.go
    dto/res_dto.go
    handler/category_handler.go
    service/category_service.go
    repository/category_repository.go
```

Ishga tushish ketma-ketligi: `helper.LoadEnv -> config.Load -> config.DBConnect -> config.RunMigrations -> config.SeedSuperadmin -> modullarni routerga ulash -> HTTP server`.

Yangi modul ham shu tuzilishda yaratiladi. Uning `Cmd(router, group, db)` funksiyasi handlerni ulaydi, `main.go` esa `Cmd`ni `/api` guruhi bilan chaqiradi.

## Qatlamlarning vazifasi

| Qism | Mas'uliyati |
| --- | --- |
| `cmd.go` | Modulning route ro'yxatdan o'tkazish kirish nuqtasi. |
| `dto/req_dto.go` | Request va filter turlari, JSON va `validate` teglari. |
| `dto/res_dto.go` | Response turlari, mavjud uslubdagi status, role va direction konstantalari. |
| Handler | Body, path va queryni o'qish; transport validatsiyasi; serviceni chaqirish; xatoni HTTP javobiga aylantirish. |
| Service | Biznes qoidalari, jarayon ketma-ketligi, repository va kerakli boshqa servicelarni chaqirish, natijani tayyorlash. |
| Repository | Parametrli SQL, `QueryRow`/`Query`/`Exec`, `Scan`, saqlash va o'qish. |
| Helper | Bir nechta joyda kerak bo'ladigan, modul biznesiga bog'lanmagan yordamchi funksiyalar. |

Handlerda matnlarni `TrimSpace` qilish, ID/queryni parse qilish va `helper.Validate` ishlatish mavjud uslubga mos. Boshqa chaqiruvchilarga ham tegishli biznes shartlari serviceda tekshiriladi. Masalan, `TournamentService.Create` va `Update` tugash vaqti boshlanish vaqtidan keyinligini tekshiradi.

Serviceda `http.ResponseWriter`, `*http.Request`, HTTP status yoki JSON yozish bo'lmasin. Repositoryda request parse qilish, JWT tekshirish yoki service chaqirish bo'lmasin. DTOlar qatlamlar o'rtasida uzatiladi; `context.Context` handlerdagi `r.Context()`dan repositorygacha yetkaziladi.

CRUD bo'lmagan amal uchun ham handler HTTP qismini, service biznes qismini, repository bazaga oid qismini bajaradi. Har yangi amal uchun yangi repository metodi shart emas: mavjud `Show`, `Update`, `FindBy...` yoki boshqa mos metod yetarli bo'lsa, qayta ishlating. Bazaga murojaat kerak bo'lmagan amal uchun sun'iy SQL yoki bo'sh repository metodi yaratmang.

## Mavjud funksiyalardan qayta foydalanish

Yozishdan oldin:

1. `rg` bilan butun `module/` va `helper/` ichida kerakli amal, tur va SQLni qidiring; faqat joriy modul bilan cheklanmang.
2. Topilgan funksiyaning imzosi, biznes tekshiruvlari, filterlari, qaytaradigan xatolari va yon ta'sirlarini o'qing.
3. Semantikasi mos bo'lsa, uni chaqiring. Mos bo'lmasa, mavjud APIga zarar bermaydigan kengaytirish yoki haqiqatan alohida funksiya kerakligini aniqlang.
4. Yangi metod zarur bo'lsa, tegishli interface va implementatsiyani birga yangilang.

Biznes tekshiruvlari bor imkoniyatni qayta ishlatishda boshqa modulning service interface'ini chaqiring. Masalan, autentifikatsiya uchun `AuthService.Authenticate`dan foydalaning; token va sessiya tekshiruvlarini boshqa modulda qayta yozmang.

Faqat ma'lumotga oid tayyor imkoniyat kerak bo'lsa va biznes tekshiruvlari chetlab o'tilmasa, service boshqa modulning repository interface'idan ham foydalanishi mumkin. Handler repositoryga bevosita chiqmaydi. Repositorylar orasida biznes jarayonini zanjirlamang; buni service boshqaradi. Aylanma importga yo'l qo'ymang.

Bog'liqlikni struct maydonida interface orqali saqlang, konstruktorda ulang. Mavjud konstruktor uslubini saqlang; yangi service yoki repositoryni har bir request ichida qayta yaratmang.

Quyidagilar hozir mavjud; ularning kodini qayta yozishdan oldin tekshiring:

| Joy | Mavjud imkoniyatlar |
| --- | --- |
| `helper/helper.go` | `LoadEnv`, `ENV`, `JSON`, `DecodeJSON`, `Validate`. `DecodeJSON` body hajmini cheklaydi, noma'lum maydon va qo'shimcha JSON qiymatlarini rad etadi. |
| `helper/http.go` | `ParseID`: route'dagi musbat `int64` IDni tekshiradi; tournament va category handlerlarida umumiy ishlatiladi. |
| `helper/auth.go` | `HashPassword`, `CheckPassword`, `NewToken`, `ParseToken`, `HashToken`, `Claims`, `TokenIssuer`. |
| `module/auth/service` | `SeedSuperadmin`, `Login`, `Authenticate`, `Logout`; ichki `tokenUserID` token tekshiruvlarida qayta ishlatiladi. |
| `module/auth/repository` | `SeedSuperadmin`, `FindByEmail`, `CreateSession`, `FindBySession`, `DeleteSession`. |
| `module/auth/handler` | Eksport qilingan `RequireAuth(service, next, roles...)`; ichki `parseToken` va `writeServiceError`. |
| `module/tournament/service` | `Create`, `List`, `Show`, `Update`, `Delete`; `ErrInvalidDates`. |
| `module/tournament/repository` | `Create`, `List`, `HasMore`, `Show`, `Update`, `Delete`. |
| `module/tournament/handler` | Modul ichida umumiy `parseBody`, `writeServiceError`; ID uchun `helper.ParseID`. |
| `module/category/service` | `Create`, `List`, `WithChildren`, `Show`, `Update`, `Delete`; turnirni tekshirishda `TournamentService.Show` qayta ishlatiladi, parent va sikl tekshiruvlari serviceda. |
| `module/category/repository` | `Create`, `List`, `WithChildren`, `Show`, `Update`, `Delete`, `WithTransaction`; `Scan` har bir metod ichida turnirdagi kabi yoziladi. `WithChildren` bosh kategoriyalar va SQLda yig'ilgan bolalar JSONini qaytaradi. |
| `module/category/handler` | `RequireAuth` va `helper.ParseID`ni qayta ishlatadi; modulga xos `parseBody`, `writeServiceError`. Filterlar `List` ichida turnirdagi kabi yoziladi. |

Kichik harf bilan boshlangan funksiyalar faqat o'z package'ida ochiq. Ularni boshqa moduldan to'g'ridan-to'g'ri chaqirish mumkin deb hisoblamang. Bir nechta modulga kerak bo'ladigan umumiy qismini zarurat bo'lganda mos umumiy joyga ajrating; tournamentga xos `parseBody`ni boshqa DTOga majburan moslashtirmang.

## Nomlash va konstruktorlar

- Package/import aliaslar: `auth_dto`, `auth_handler`, `auth_service`, `auth_repository`, `tournament_dto` kabi `<module>_<layer>` uslubi.
- Interfacelar eksport qilinadi: `TournamentService`, `TournamentRepository`. Implementatsiyalar yopiq: `tournamentService`, `tournamentRepository`, `tournamentHandler`.
- Receiver nomlari: `handler`, `service`, `repository`.
- Konstruktorlar: `NewTournamentHandler(router, group, db)`, `NewTournamentService(db) TournamentService`, `NewTournamentRepository(db) TournamentRepository`.
- Handler o'z service interface'ini, service repository interface'ini, repository `*pgxpool.Pool`ni saqlaydi.
- Mavjud CRUD nomlari: `Create`, `List`, `Show`, `Update`, `Delete`. Boshqa amallarga maqsadini tushuntiradigan nom bering; ularni sun'iy ravishda CRUD nomlariga aylantirmang.
- Mavjud role/status/direction konstantalarini qayta ishlating; bir xil qiymat uchun turli joylarda yangi konstantalar yaratmang.

## Qavslar va mantiqiy bloklar uslubi

`module/tournament/handler/tournament_handler.go`, `service/tournament_service.go` va `repository/tournament_repository.go`ni namuna deb oling. Ochiluvchi `{` funksiya, `if`, `for` bilan bir qatorda yoziladi. Alohida mantiqiy blok esa keyingi qatorda ochiladi. Tab bilan indentatsiya qiling, `gofmt` ishlating.

Handler va servicedagi mantiqiy tekshiruvlar uchun mavjud alohida `{ ... }` bloklar uslubini saqlang. Repository SQL chaqiruvidan keyin esa oddiy `if err != nil` ishlating. Handler uchun namuna:

```go
res, err := handler.service.Create(r.Context(), &req)
{
	if err != nil {
		writeServiceError(w, err)
		return
	}
}

helper.JSON(w, http.StatusCreated, res)
```

Request parse qilish ham shu uslubda:

```go
var req tournament_dto.Create
{
	if !parseBody(w, r, &req) {
		return
	}
}
```

Route'larni umumiy prefiksdan keyingi blokda guruhlang:

```go
routes := group + "/tournaments"
{
	router.POST(routes, auth_handler.RequireAuth(auth, handler.Create, roles...))
	router.GET(routes, auth_handler.RequireAuth(auth, handler.List))
}
```

Xato yoki noto'g'ri holatda erta `return` qiling. Mantiqiy bosqichlarni bo'sh qator bilan ajrating. Mavjud kodda oddiy `if` va to'g'ridan-to'g'ri `return service.repository.Show(ctx, id)` ham bor; oddiy delegatsiyani keraksiz o'zgaruvchi yoki bloklar bilan murakkablashtirmang. Mavjud alohida bloklarni faqat qisqartirish uchun olib tashlamang. Qo'shni kodning formatini keraksiz qayta yozmang.

## HTTP, xatolar va ma'lumotlar bazasi

- JSON javobi uchun `helper.JSON`, body uchun `helper.DecodeJSON`, DTO tekshiruvi uchun `helper.Validate`dan foydalaning.
- Mavjud javob shakllari: oddiy xato `{"error": "..."}`, validatsiya xatolari `{"errors": {...}}`.
- Service xatolarini modulning `writeServiceError` funksiyasida HTTP statuslarga aylantiring; `errors.Is`/`errors.As`dan foydalaning.
- Tournamentda `pgx.ErrNoRows -> 404`, `ErrInvalidDates -> 400`; kutilmagan xatolar log qilinadi va mijozga umumiy `500` javobi beriladi. Authning mavjud `401`/`403` xaritalashini saqlang.
- Yaratishda `201`, muvaffaqiyatli o'qish/yangilashda `200`, o'chirish yoki logoutda bodysiz `204` mavjud uslubdir.
- Himoyalangan yangi route'larda `auth_handler.RequireAuth`ni qayta ishlating. Hozir tournament yozish amallari `superadmin`/`admin`, o'qish amallari autentifikatsiyadan o'tgan foydalanuvchilar uchun ochiq.
- SQLni repositoryda saqlang, foydalanuvchi qiymatlarini `$1`, `$2` kabi parametrlar bilan uzating. Tournamentdagi dinamik operator va tartib faqat koddagi belgilangan qiymatlardan tanlanadi; foydalanuvchi matnini SQLga qo'shmang.
- `Query`dan keyin `defer rows.Close()`, iteratsiyadan keyin `rows.Err()` tekshiruvini saqlang. `Exec`da yozuv topilmaganini aniqlash kerak bo'lsa, `RowsAffected()`ni tekshiring.
- Bir nechta yozuvni atomar o'zgartirish zarur bo'lsa, jarayonni service boshqarsin, SQL esa repositoryda qolsin; tegishli operatsiyalar bir tranzaksiyada bajarilsin.
- Schema o'zgarishini `config/migrations/` ichidagi keyingi raqamli SQL faylida yozing. Migratsiyalarni qo'llaydigan mavjud mexanizmni qayta ishlating.

Tournament ro'yxati cursor asosida ishlaydi: repository `Limit+1` yozuv oladi, service ortiqcha yozuvni ajratadi, `prev` natijasini teskari aylantiradi va `HasMore` orqali qarama-qarshi yo'nalishni tekshiradi. Shu imkoniyatni qayta yozishdan oldin `List` va `HasMore`ni o'qing. Hozir HTTPda standart limit `10`, serviceda yaroqsiz limit uchun fallback `20`; mavjud farqni vazifasiz o'zgartirmang.

Category API `/api/categories`da. `parent_id: null` bosh kategoriyani bildiradi; parent shu turnirga tegishli bo'lishi va sikl yaratmasligi kerak. Service tekshiruv va yozishni repositoryning `WithTransaction` metodi orqali serializable tranzaksiyada bajaradi. `tournament_id` FK va `(parent_id, tournament_id)` kompozit FK bog'lanishlarni bazada ham himoya qiladi. Ichki kategoriyasi bor kategoriyani o'chirish yoki boshqa turnirga ko'chirish `409`; kategoriyalari bor turnirni o'chirish ham `409`. Category ro'yxatida oldinga cursor pagination va `tournament_id`, `parent_id`, `name`, `is_active` filterlari bor. `parent_id` yuborilmasa yoki `parent_id=0` bo'lsa, faqat `parent_id IS NULL` bosh kategoriyalar qaytadi; musbat ID yuborilsa, shu parentning ichki kategoriyalari qaytadi.

`GET /api/categories_with_childr?tournament_id=1` turnirning kategoriyalar daraxtini qaytaradigan alohida API. Category modulining route blokida `WithChildren` handleriga to'g'ridan-to'g'ri ulangan. `Show` faqat `/api/categories/:id` uchun ishlaydi. `tournament_id` majburiy musbat ID; mavjud bo'lmagan turnir `404`, kategoriya yo'q bo'lsa `{"data":[]}`. Service faqat mavjud turnirni tekshirib, repository natijasini qaytaradi; daraxt yig'uvchi `for`, map yoki `json.Marshal` serviceda bo'lmasin. Repository `WithChildren` SQL orqali shu turnirdagi `parent_id IS NULL` kategoriyalarni oladi. `004_create_category_children_function.sql` migratsiyasidagi rekursiv `category_children_json` funksiyasi ularning barcha darajadagi bolalarini SQLda ichma-ich JSON massivga yig'adi. Root va bolalar ID bo'yicha kamayish tartibida. `WithChildrenResponse.Children` turi `json.RawMessage`; repository SQLdan tayyor JSONni shu maydonga `Scan` qiladi. Bolasi yo'q kategoriyada `children: []`. Pagination va faol holat filtri qo'llanmaydi. Mavjud auth tekshiruvi saqlanadi.

## O'zgartirishni yakunlash

1. O'zgargan interface, implementatsiya, DTO, route va migratsiyalar o'zaro mosligini tekshiring.
2. O'zgargan Go fayllariga `gofmt -w <fayllar>` ishlating.
3. Go kodi o'zgarsa kompilyatsiya va `go vet ./...` bilan tekshiring; zarur bo'lsa amaldagi xulqni vaqtinchalik tekshiring. Test fayllarini loyihada qoldirmang, `.lock` fayllarini yaratmang. Faqat hujjat o'zgarsa kod tekshiruvlarini qayta ishga tushirish shart emas.
4. Tekshiruvni bajarib bo'lmasa, sababini aniq ayting; o'tmagan tekshiruvni o'tdi demang.
5. `go run .` bazaga ulanib migratsiya va seederni ishga tushirishini hisobga oling; uni oddiy kompilyatsiya tekshiruvi sifatida ishlatmang.
6. Ushbu xarita eskirmasligi uchun tuzilish yoki qayta ishlatiladigan API o'zgarganda hujjatni ham yangilang.

