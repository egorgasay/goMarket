# goMarket
# LIVE DEMO: https://gomarket.egorgasay.repl.co
## Template for online store with loyalty system and admin page. Microservice architecture.

## Interaction scheme
![изображение](https://user-images.githubusercontent.com/102957432/224708312-e50c521b-4448-4008-a73f-0d02788cc5d2.png)

## Launch 
```bash
docker-compose up -d
```

### Loyalty Mechanics
Flexible model of interaction with the loyalty system. In the example, 10% of each purchase of a product with the letter "B" in the product name is credited. You can change this behavior by adding new mechanics to the accrual service.

### Example
```http
POST http://ACCRUAL_ADDRESS/api/goods HTTP/1.1
Content-Type: application/json

{
    "match": "Bork",
    "reward": 10,
    "reward_type": "%"
} 
```

# Overview:  
## Main page   
![изображение](https://user-images.githubusercontent.com/102957432/224706670-26694b8e-7ffb-48b5-8865-555193dabbe5.png)

## Basket  
![изображение](https://user-images.githubusercontent.com/102957432/224706773-8f3d24d6-779e-45b9-a5cc-dde90f2e065f.png)

## Order proccess  
![изображение](https://user-images.githubusercontent.com/102957432/224709001-f8b9aa5b-89eb-4c16-8da0-c023cf22beb3.png)


## The page with the receipt
![изображение](https://user-images.githubusercontent.com/102957432/224706957-f579ad27-44bc-4839-a67d-53ab3a05b085.png)

# Admin page
## Orders  
![изображение](https://user-images.githubusercontent.com/102957432/224707166-5720097a-69df-46c9-87e7-f577e140688c.png)

## Edit items
![изображение](https://user-images.githubusercontent.com/102957432/224707325-495da616-f0b1-4a59-88e7-c1cda2e707ac.png)

## Add an item  
![изображение](https://user-images.githubusercontent.com/102957432/224707358-a6491ff1-0bae-4f19-bc6d-c4ea9e5738ca.png)
