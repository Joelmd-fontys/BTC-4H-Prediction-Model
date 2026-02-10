CREATE VIEW IF NOT EXISTS dataset AS
SELECT
    f.exchange,
    f.symbol,
    f.timeframe,
    f.timestamp,

    -- features
    f.ret_1,
    f.vol_20,
    f.mom_6,
    f.ema_spread,
    f.range_hl,
    f.range_co,
    f.vol_chg,

    -- labels
    l.label,
    l.fwd_ret,
    l.threshold_b

FROM features f
         JOIN labels l
              ON f.exchange = l.exchange
                  AND f.symbol = l.symbol
                  AND f.timeframe = l.timeframe
                  AND f.timestamp = l.timestamp

WHERE
    f.ret_1 IS NOT NULL
  AND f.vol_20 IS NOT NULL
  AND f.mom_6 IS NOT NULL
  AND f.ema_spread IS NOT NULL
  AND f.range_hl IS NOT NULL
  AND f.range_co IS NOT NULL
  AND f.vol_chg IS NOT NULL;