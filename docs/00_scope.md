# Scope (v0.1) — BTC 4H Quantitative Direction Pipeline

## Purpose
Build a reproducible, API-driven research pipeline that predicts *meaningful* directional moves in Bitcoin over a 4-hour horizon and evaluates whether those predictions translate into a risk-aware paper-traded strategy.

This project is educational and analytical. It is not financial advice and will not place live trades.

---

## Core Deliverable (What we are building)
A deterministic pipeline that can be run end-to-end from the command line:

1) Fetch 4H BTC/USDT spot candles from Binance REST API
2) Validate and store candles with time-series integrity checks
3) Compute features using only past information
4) Generate 3-class labels (UP / DOWN / NO_TRADE) based on a thresholded next-candle log return
5) Train baseline predictive models (starting with multinomial logistic regression)
6) Evaluate with strict walk-forward (time-ordered) splits
7) Simulate a paper strategy (with fees/slippage, risk controls)
8) Generate reports (metrics, diagnostics, and equity curve)

Primary outputs:
- Local database of candles/features/labels
- Model predictions (probabilities) per timestamp
- Walk-forward evaluation metrics
- Paper trading performance report

---

## Prediction Target (Frozen)
- Asset: BTC/USDT (Binance Spot)
- Timeframe: 4-hour candles (4H)
- Target: next-candle log return  
  r_{t+1} = ln(P_{t+1} / P_t), where P_t is the 4H close at time t

3-class label with threshold b:
- UP if r_{t+1} > +b
- DOWN if r_{t+1} < -b
- NO_TRADE otherwise

Threshold options supported in v0.1:
- Fixed b (initial default in config, e.g. 0.6%)
- Volatility-adjusted b_t = k * rolling_volatility(t) where rolling volatility is computed from past returns only

---

## Evaluation Protocol (Frozen)
- All training/testing must be time-ordered (no random splits).
- Walk-forward evaluation is required (expanding or rolling windows).
- No future information may be used in feature computation, label thresholds, scaling/normalization, or strategy rules.
- Baselines must be included and reported:
    - Always NO_TRADE
    - Random with class priors
    - Simple heuristic (e.g., sign of previous return)

Primary evaluation metrics:
- Confusion matrix (3-class)
- Precision/recall for UP and DOWN
- Action rate (fraction of predicted UP/DOWN)
- Calibration diagnostics for class probabilities
- Strategy metrics: cumulative return, max drawdown, turnover, trade count, Sharpe (with clear assumptions)

---

## Engineering Constraints (Non-Negotiable)
- Determinism: the same inputs/config produce the same outputs.
- Data integrity: detect and handle missing/duplicate/out-of-order candles.
- No rate-limit evasion: follow Binance API constraints.
- Reproducibility: configs and run metadata are logged/snapshotted.
- Clean separation of concerns: exchange client, storage, features, modeling, evaluation, reporting.

---

## In Scope (v0.1)
Data:
- Binance Spot REST candles (klines)
- Local storage (SQLite initially)

Features (v0.1):
- Log returns (1 and multi-bar)
- Rolling volatility (multiple windows)
- Momentum over k bars
- EMA spreads (short vs long)
- Range features (high-low, close-open)
- Volume change

Models (v0.1):
- Baselines listed above
- Multinomial logistic regression with regularization

Backtest / paper strategy:
- Trades only when model confidence exceeds a threshold
- Simple position sizing (confidence-based, optional volatility scaling)
- Fees + slippage modeled conservatively
- Enter/exit rules aligned to 4H horizon (defined explicitly in evaluation protocol)

Reporting:
- Fold-by-fold walk-forward metrics
- Calibration diagnostics
- Equity curve + drawdown
- Confidence bucket analysis

---

## Out of Scope (Explicit Non-Goals)
- Live trading or exchange order execution
- High-frequency / order book microstructure
- Deep learning / neural networks (v0.1)
- Arbitrage strategies
- Alternative data (on-chain, funding rates, social, etc.)
- Optimizing hyperparameters to maximize PnL on the full dataset
- Multi-exchange aggregation or cross-venue routing

---

## Definition of Success
The project is successful if:
- The pipeline is correct, clean, and reproducible
- Results are honest and out-of-sample (walk-forward)
- Limitations are clearly documented
- The system demonstrates *when not to trade* as well as when to trade

A weak or negative result is still a valid outcome.

---

## Versioning Notes
- v0.1 focuses on correctness and evaluation harness quality.
- Additional complexity (more features, regime models, alternative data, advanced models) is considered only after v0.1 is stable.