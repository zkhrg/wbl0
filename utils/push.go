package main

import (
	stan "github.com/nats-io/stan.go"
)

var valid_json1 = `{
"order_uid": "test111",
"track_number": "dsfgdsfg",
"entry": "WBIL",
"delivery": {
    "name": "Test2 Testov2",
    "phone": "+9720000000",
    "zip": "2639809",
    "city": "Kiryat Mozkin",
    "address": "Ploshad Mira 15",
    "region": "Kraiot",
    "email": "test@gmail.com"
},
"payment": {
    "transaction": "test111",
    "request_id": "",
    "currency": "USD",
    "provider": "wbpay",
    "amount": 1817,
    "payment_dt": 1637907727,
    "bank": "alpha",
    "delivery_cost": 1500,
    "goods_total": 317,
    "custom_fee": 0
},
"items": [
    {
    "chrt_id": 9934930,
    "track_number": "dsfgdsfg",
    "price": 453,
    "rid": "ab4219087a764ae0btest",
    "name": "Mascaras",
    "sale": 30,
    "size": "0",
    "total_price": 317,
    "nm_id": 2389212,
    "brand": "Vivienne Sabo",
    "status": 202
    }
],
"locale": "en",
"internal_signature": "",
"customer_id": "test",
"delivery_service": "meest",
"shardkey": "9",
"sm_id": 99,
"date_created": "2021-11-26T06:22:19Z",
"oof_shard": "1"
}`
var test1 = `{
  "order_uid": "b563feb7b2b84b6testasddd",
  "track_number": "WB",
  "entry": "WBIL",
  "delivery": {
    "name": "Test Testov",
    "phone": "+9720000000",
    "zip": "2639809",
    "city": "Kiryat Mozkin",
    "address": "Ploshad Mira 175",
    "region": "Kraiot",
    "email": "test@gmail.com"
  },
  "payment": {
    "transaction": "b563feb7b2b84b6testasddd",
    "request_id": "",
    "currency": "USD",
    "provider": "wbpay",
    "amount": 1817,
    "payment_dt": 1637907727,
    "bank": "alpha",
    "delivery_cost": 1500,
    "goods_total": 317,
    "custom_fee": 0
  },
  "items": [
    {
      "chrt_id": 9934930,
      "track_number": "WB",
      "price": 453,
      "rid": "rthrh77",
      "name": "Mascaras",
      "sale": 30,
      "size": "0",
      "total_price": 317,
      "nm_id": 2389212,
      "brand": "Vivienne Sabo",
      "status": 202
    }
  ],
  "locale": "en",
  "internal_signature": "",
  "customer_id": "test",
  "delivery_service": "meest",
  "shardkey": "9",
  "sm_id": 99,
  "date_created": "2021-11-26T06:22:19Z",
  "oof_shard": "1"
}
`
var many_items = `{
"order_uid": "manyitems",
"track_number": "WBILMTESTTRACK",
"entry": "WBIL",
"delivery": {
    "name": "Test Testov",
    "phone": "+9720000000",
    "zip": "2639809",
    "city": "Kiryat Mozkin",
    "address": "Ploshad Mira 15",
    "region": "Kraiot",
    "email": "test@gmail.com"
},
"payment": {
    "transaction": "manyitems",
    "request_id": "",
    "currency": "USD",
    "provider": "wbpay",
    "amount": 1817,
    "payment_dt": 1637907727,
    "bank": "alpha",
    "delivery_cost": 1500,
    "goods_total": 317,
    "custom_fee": 0
},
"items": [
    {
    "chrt_id": 1,
    "track_number": "WBILMTESTTRACK",
    "price": 453,
    "rid": "ab4219087a764ae0btest",
    "name": "whatever",
    "sale": 30,
    "size": "0",
    "total_price": 317,
    "nm_id": 2389212,
    "brand": "Vivienne Sabo",
    "status": 202
    },
    {
    "chrt_id": 6213,
    "track_number": "WBILMTESTTRACK",
    "price": 453,
    "rid": "ab4219087a764ae0btest",
    "name": "WhiteHat",
    "sale": 30,
    "size": "0",
    "total_price": 317,
    "nm_id": 2389212,
    "brand": "Vivienne Sabo",
    "status": 202
    },
    {
    "chrt_id": 431235555,
    "track_number": "WBILMTESTTRACK",
    "price": 453,
    "rid": "ab4219087a764ae0btest",
    "name": "RedHat",
    "sale": 30,
    "size": "0",
    "total_price": 317,
    "nm_id": 2389212,
    "brand": "Vivienne Sabo",
    "status": 202
    },
    {
    "chrt_id": 2222203,
    "track_number": "WBILMTESTTRACK",
    "price": 453,
    "rid": "ab4219087a764ae0btest",
    "name": "BlackHat",
    "sale": 30,
    "size": "0",
    "total_price": 317,
    "nm_id": 2389212,
    "brand": "Vivienne Sabo",
    "status": 202
    }
],
"locale": "en",
"internal_signature": "",
"customer_id": "test",
"delivery_service": "meest",
"shardkey": "9",
"sm_id": 99,
"date_created": "2021-11-26T06:22:19Z",
"oof_shard": "1"
}`
var valid_json2 = `{
"order_uid": "op345345csdlla",
"track_number": "test_track_number",
"entry": "WBIL",
"delivery": {
    "name": "Test Testov",
    "phone": "+9720800000",
    "zip": "2639809",
    "city": "Nobel Mozkin",
    "address": "Ploshad Mira 15",
    "region": "Kraiot",
    "email": "test-set@gmail.com"
},
"payment": {
    "transaction": "op345345csdlla",
    "request_id": "",
    "currency": "USD",
    "provider": "wbpay",
    "amount": 7447,
    "payment_dt": 1637907727,
    "bank": "omega",
    "delivery_cost": 1500,
    "goods_total": 317,
    "custom_fee": 0
},
"items": [
    {
    "chrt_id": 9934930,
    "track_number": "test_track_number",
    "price": 453,
    "rid": "ghfgh678",
    "name": "Mascaras",
    "sale": 30,
    "size": "0",
    "total_price": 317,
    "nm_id": 2389212,
    "brand": "Vivienne Sabo",
    "status": 202
    }
],
"locale": "en",
"internal_signature": "",
"customer_id": "test",
"delivery_service": "meest",
"shardkey": "9",
"sm_id": 99,
"date_created": "2021-11-26T06:22:19Z",
"oof_shard": "1"
}`

func main() {
	sc, _ := stan.Connect("test-cluster", "s", stan.NatsURL("nats://localhost:4223"))
	sc.Publish("service", []byte(valid_json1))
	sc.Publish("service", []byte("{invalid}"))
	sc.Publish("service", []byte(test1))
	sc.Publish("service", []byte(many_items))
	sc.Publish("service", []byte(valid_json2))
	sc.Publish("service", []byte("invalid_data+another_invalid_data"))
}
