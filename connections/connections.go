package connections

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"wb-level0/models"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/joho/godotenv"
	"github.com/nats-io/stan.go"
	"github.com/patrickmn/go-cache"
)

func DataBase() *pgx.Conn {
	fmt.Println("Init app..")
	fmt.Print("Init environment..")
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Not found .env file")
		os.Exit(1)
	}
	fmt.Print("complete\n")
	fmt.Printf("Conn to bd: %v ... ", os.Getenv("DB_NAME"))
	urlExample := fmt.Sprintf("%v://%v:%v@%v:%v/%v", os.Getenv("DRIVER"), os.Getenv("USERNAME"),
		os.Getenv("PASSWORD"), os.Getenv("HOST"), os.Getenv("PORT"), os.Getenv("DB_NAME"))
	conn, err := pgx.Connect(context.Background(), urlExample)
	if err != nil {
		fmt.Printf("Failed conn to bd: %v\n", err)
		os.Exit(1)
	}
	fmt.Print("complete\n")

	return conn
}

func CreateCache(conn *pgx.Conn) *cache.Cache {
	order_uid := GetOrderUid(conn)
	Cache := cache.New(-1, -1)
	for i := range order_uid {
		data, _ := GetDataByUid(conn, order_uid[i])
		Cache.Set(order_uid[i], data, cache.NoExpiration)
	}
	fmt.Printf("Восстановление базы данных в кэш. Записей обнаружено: (%v)\n", len(Cache.Items()))

	return Cache
}

func NatsStreaming(conn *pgx.Conn, Cache *cache.Cache) *stan.Conn {
	// connect to nats-streaming-service
	sc, _ := stan.Connect("test-cluster", "sm", stan.NatsURL("nats://localhost:4223"))

	var order models.OrderInfo
	// подписка на канал для дальнейшей обработки полученных данных
	sc.Subscribe("service", func(m *stan.Msg) {
		err := json.Unmarshal(m.Data, &order)
		if err != nil {
			fmt.Println("Got: Invalid json")
			InsertInvalidData(conn, string(m.Data))
		} else {
			fmt.Println("Got: Valid json")
			Cache.Set(order.OrderUid, string(m.Data), cache.NoExpiration)
			InsertData(conn, order)
		}
	})

	fmt.Println("Успешно подписаны!")
	return &sc
}

func RunHttpServer(Cache *cache.Cache) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	fmt.Println("Http server runs at http://localhost:8080")
	r.LoadHTMLGlob("templates/*.html")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"content": "This is an index page...",
		})
	})
	r.POST("/result", func(c *gin.Context) {
		result, _ := Cache.Get(c.PostForm("order_uid"))
		if result == nil {
			c.PureJSON(http.StatusOK, "Not found this order_uid")
		} else {
			c.PureJSON(http.StatusOK, result)
		}

	})
	r.Run(":8080")
}

func GetOrderUid(conn *pgx.Conn) (slice_uid []string) {

	query := `
		select array_agg(order_uid) from order_info		
	`
	err := conn.QueryRow(context.Background(), query).Scan(&slice_uid)
	if err != nil {
		fmt.Println(err)
	}
	return slice_uid
}

func GetDataByUid(conn *pgx.Conn, order_uid string) (string, error) {
	var order models.OrderInfo
	query := `
		select oi.*, to_jsonb(p.*) as "payment", (select jsonb_agg((to_jsonb(i.*))) from items i )  as "items",
			to_jsonb(del) as "delivery" 
		from order_info oi 
		left join payments p on p."transaction" = oi.order_uid 
		left join items i on i.track_number = oi.track_number
		join (
				select d.name,d.phone ,d.zip ,d.city ,d.address ,d.region ,d.email  
				from deliveries d
				where d.id = (
								select od.delivery_id 
								from order_delivery od 
								where od.order_uid=$1
								)
			) as del on true
		where oi.order_uid = $1
		limit 1`
	if err := conn.QueryRow(context.Background(), query, order_uid).Scan(
		&order.OrderUid,
		&order.TrackNumber,
		&order.Entry,
		&order.Locale,
		&order.InternalSignature,
		&order.CustomerId,
		&order.DeliveryService,
		&order.Shardkey,
		&order.SmId,
		&order.DateCreated,
		&order.OofShard,
		&order.Payment,
		&order.Items,
		&order.Delivery,
	); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			fmt.Println(pgErr)
			return "", pgErr
		}
	}
	res, _ := json.Marshal(&order)
	return string(res), nil
}

func InsertInvalidData(conn *pgx.Conn, data string) (err error) {
	query := `insert into invalid_data(data) values ($1)`
	if err = conn.QueryRow(context.Background(), query, data).Scan(); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			fmt.Println(pgErr)
			return pgErr
		}

	}
	return nil
}

func InsertData(conn *pgx.Conn, order models.OrderInfo) {
	err := InsertDataPayment(conn, order)
	if err != nil {
		fmt.Println(err)
	}
	uid, err := InsertDataOrder(conn, order)
	if err != nil {
		fmt.Println(err)
	}
	id, err := InsertDataDelivery(conn, order)
	if err != nil {
		fmt.Println(err)
	}
	err = InsertOrderDelivery(conn, uid, id)
	if err != nil {
		fmt.Println(err)
	}
	err = InsertDataItems(conn, order)
	if err != nil {
		fmt.Println(err)
	}
}

func InsertDataOrder(conn *pgx.Conn, order models.OrderInfo) (order_uid string, err error) {
	query := `
		insert into order_info
			(order_uid, track_number, entry, locale, internal_signature,customer_id, delivery_service, shardkey, sm_id, date_created,oof_shard) 
		values 
			($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		returning order_uid
		`
	if err = conn.QueryRow(context.Background(), query,
		order.OrderUid,
		order.TrackNumber,
		order.Entry,
		order.Locale,
		order.InternalSignature,
		order.CustomerId,
		order.DeliveryService,
		order.Shardkey,
		order.SmId,
		order.DateCreated,
		order.OofShard,
	).Scan(&order_uid); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			return "", pgErr
		}
	}
	return order_uid, nil
}

// Insert delivery into db
func InsertDataDelivery(conn *pgx.Conn, order models.OrderInfo) (id int, err error) {
	query := `
		insert into deliveries
			(name, phone, zip, city, address, region, email) 
		values 
			($1,$2,$3,$4,$5,$6,$7)
		returning id
		`
	if err = conn.QueryRow(context.Background(), query,
		order.Delivery.Name,
		order.Delivery.Phone,
		order.Delivery.Zip,
		order.Delivery.City,
		order.Delivery.Address,
		order.Delivery.Region,
		order.Delivery.Email,
	).Scan(&id); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			fmt.Println(pgErr)
			return 0, pgErr
		}
	}
	return id, nil
}

// Insert payments into db
func InsertDataPayment(conn *pgx.Conn, order models.OrderInfo) (err error) {
	query := `
		insert into payments
			(transaction, request_id, currency, provider, amount, payment_dt, bank, delivery_cost, goods_total,
                      custom_fee) 
		values 
			($1,$2,$3,$4,$5,$6,$7, $8, $9, $10)
		`
	if err = conn.QueryRow(context.Background(), query,
		order.Payment.Transaction,
		order.Payment.RequestId,
		order.Payment.Currency,
		order.Payment.Provider,
		order.Payment.Amount,
		order.Payment.PaymentDt,
		order.Payment.Bank,
		order.Payment.DeliveryCost,
		order.Payment.GoodsTotal,
		order.Payment.CustomFee,
	).Scan(); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			fmt.Println(pgErr)
			return pgErr
		}
	}
	return nil
}

// insert order_delivery into db
func InsertOrderDelivery(conn *pgx.Conn, order_uid string, id int) (err error) {
	query := `
		insert into order_delivery
			(order_uid,delivery_id)
		values 
			($1,$2)
		`
	if err = conn.QueryRow(context.Background(), query, order_uid, id).Scan(); err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			fmt.Println(pgErr)
			return pgErr
		}
	}
	return nil
}

// insert Items into db
func InsertDataItems(conn *pgx.Conn, order models.OrderInfo) (err error) {
	for i := 0; i < len(order.Items); i++ {
		query := `
			insert into items
				(chrt_id,track_number,price,rid,name,sale,size,total_price,nm_id,brand,status)
			values
				($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
			`
		if err = conn.QueryRow(context.Background(), query,
			order.Items[i].ChrtId,
			order.Items[i].TrackNumber,
			order.Items[i].Price,
			order.Items[i].Rid,
			order.Items[i].Name,
			order.Items[i].Sale,
			order.Items[i].Size,
			order.Items[i].TotalPrice,
			order.Items[i].NmId,
			order.Items[i].Brand,
			order.Items[i].Status,
		).Scan(); err != nil {
			if pgErr, ok := err.(*pgconn.PgError); ok {
				fmt.Println(pgErr)
				return pgErr
			}
		}
	}

	return nil
}
