CREATE TABLE IF NOT EXISTS predictions (
  exchange   TEXT NOT NULL,
  symbol     TEXT NOT NULL,
  timeframe  TEXT NOT NULL,
  timestamp  INTEGER NOT NULL,

  model_name TEXT NOT NULL,

  p_up       REAL NOT NULL,
  p_down     REAL NOT NULL,
  p_no_trade REAL NOT NULL,

  predicted_label TEXT NOT NULL,
  actual_label    TEXT NOT NULL,

  PRIMARY KEY (exchange, symbol, timeframe, timestamp, model_name),
  CHECK (predicted_label IN ('UP','DOWN','NO_TRADE')),
  CHECK (actual_label IN ('UP','DOWN','NO_TRADE'))
);