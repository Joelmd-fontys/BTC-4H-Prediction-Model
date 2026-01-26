# Data Contract: Candles (4H)

## Timeframe
- Candle duration: 4 hours (4H)
- Time alignment: Binance-defined boundaries (as returned by the API)
- Expected interval between candles: 14,400,000 ms (4 hours)

## Timestamp Semantics
- Stored timestamp `ts` represents: **open time** of the 4H candle
- Units: Unix epoch **milliseconds**
- Timezone: **UTC** (epoch timestamps)

## OHLCV Fields
All price fields are floats/decimals as returned by Binance:
- open: opening price for the 4H interval
- high: highest price during the interval
- low: lowest price during the interval
- close: closing price for the interval
- volume: base asset volume during the interval

## Invariants (Validation Rules)
For each candle row:
- high >= max(open, close)
- low <= min(open, close)
- low <= high
- volume >= 0

For the full series:
- `ts` values are strictly increasing
- Consecutive candles should differ by exactly 14,400,000 ms
    - If not, a gap is recorded as missing data

## Missing / Partial Candles
- Missing candles: allowed to exist historically; must be detected and logged.
- Partial (in-progress) candle:
    - v0.1 rule: **store only closed candles**
    - During ingestion, if the most recent returned candle is not yet closed, it is excluded.

## Source of Truth
- Exchange: Binance Spot
- Endpoint: `GET /api/v3/klines`
- Interval parameter: `interval=4h`
- Symbol: `BTCUSDT`
- Kline identity: candles are keyed by **open time** (openTime)
- pKey: exchange, symbol, timeframe, ts