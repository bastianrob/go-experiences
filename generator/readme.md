# Exploring Generator Pattern in Go

> DDD, CQRS, Actor Model, State Machine, Saga, Aggregate Root
> 
> Generator Pattern, Yield, Lazy, Channel, CSM, Worker


Imagine a `Domain` event called `Order.Placed`. Each `Order.Placed` contains:

```json
{
    "customer": "CUST-001",
    "merchant": "MRCN-001",
    "payment": "CARD-001",
    "items": [
        {"id": "IT-001", "qty": 1},
        {"id": "IT-002", "qty": 2}
    ],
    "promo": "PROM-001"
}
```

Everytime an `order` is placed, there must be something in the `Backend` service that:
* Getting customer's detail
* Getting merchant's detail
* Getting payment's detail
* Getting promotion's detail
* Getting each product / item detail
* Calculate `sum` of the order and deduct it with promotion
* Make an `order` based on the gathere information, and notify it to `merchant`
* Make an `invoice` based on the gathered information, and notify it to `customer`
* Make a payment through `payment` service based on the `invoice`
* Flag the invoice `state` as `paid`
* Notify merchant that `payment` is done to said `order`
* Notify customer that `payment` success to said `invoice`

That's **Tremendously Ridiculous** amount of work needed to be done in the `Backend` side.
It sounds `DUMB` to just make end `API` endpoint to handle these work, especially if your `Platform` can receive a ludicrous amount of order per minute across multiple region.

## Aggregate Root as Actor

First of all, let's take on the rate of `order` issue. 
`Order` can be placed at any time by millions of `customer` at the same time and the system must NEVER lose any of it.
System can defer processing an order by placing it in a `queue` which can be persisted and recovered in case of server restart, crash, or whatever disaster that comes.