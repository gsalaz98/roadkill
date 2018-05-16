We will command a Command and Control-based architecture to broker messages from seperate goroutines
around the world. We need to work on a way to do arbitrage strategies with the lowest latency.


```
+---------+                            +-----------+
| BitMEX  | <======> CNC <============>| Bitstamp  |
+---------+                            +-----------+

CNC will run our trading strategy by brokering messages from and to these two exchanges.
```