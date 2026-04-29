---
name: data-analysis
description: >
  Analyze datasets, generate charts, and create summary reports from CSV, Excel,
  JSON, Parquet, or other tabular data. Capabilities: statistical profiling,
  outlier detection, pivot tables, groupby aggregation, time-series analysis,
  correlation matrices, and publication-ready visualizations.
  Use when the user asks to analyze, visualize, profile, or summarize tabular
  data, or mentions CSV, Excel, charts, statistics, EDA, histograms, scatter
  plots, dashboards, or exploratory analysis.
---

# Data Analysis

## Workflow

Follow these steps in order. Do not skip validation checkpoints.

### Step 1: Load and validate

Select the reader based on file format:

| Format | Reader |
|--------|--------|
| CSV | `pd.read_csv(path, parse_dates=True)` |
| Excel | `pd.read_excel(path, sheet_name=0)` |
| JSON | `pd.read_json(path)` |
| Parquet | `pd.read_parquet(path)` |

For CSV files, wrap in a try/except to handle encoding issues (fall back to `encoding='latin-1'`).

```python
import pandas as pd

try:
    df = pd.read_csv('data.csv', parse_dates=True)
except UnicodeDecodeError:
    df = pd.read_csv('data.csv', encoding='latin-1', parse_dates=True)

assert not df.empty, "Dataset is empty — verify the file path and format."
print(f"Shape: {df.shape}\nColumns: {list(df.columns)}\nDtypes:\n{df.dtypes}")
```

### Step 2: Profile and clean

```python
print(df.describe(include='all'))

missing = df.isnull().sum()
print(f"Missing values:\n{missing[missing > 0]}")

for col in df.select_dtypes('object'):
    converted = pd.to_numeric(df[col], errors='coerce')
    if converted.notna().sum() > len(df) * 0.5:
        coerced = df[col][converted.isna() & df[col].notna()]
        print(f"Coerced {len(coerced)} non-numeric values in '{col}'")
        df[col] = converted

for col in df.select_dtypes('number'):
    q1, q3 = df[col].quantile([0.25, 0.75])
    iqr = q3 - q1
    outliers = ((df[col] < q1 - 1.5 * iqr) | (df[col] > q3 + 1.5 * iqr)).sum()
    if outliers > 0:
        print(f"Column '{col}': {outliers} outliers detected")
```

### Step 3: Transform and aggregate

```python
summary = df.groupby('category')['value'].agg(['mean', 'median', 'std', 'count'])
print(summary)

pivot = df.pivot_table(values='revenue', index='region', columns='quarter', aggfunc='sum')

ts = df.set_index('date')['value'].resample('M').mean()  # time-series resampling
```

### Step 4: Visualize and save

```python
import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt
import seaborn as sns

fig, axes = plt.subplots(2, 2, figsize=(14, 10))
df['value'].hist(ax=axes[0, 0], bins=30, edgecolor='black')
axes[0, 0].set_title('Distribution of Value')
df.groupby('region')['revenue'].sum().plot.bar(ax=axes[0, 1], title='Revenue by Region')
sns.heatmap(df.select_dtypes('number').corr(), annot=True, fmt='.2f', ax=axes[1, 0])
axes[1, 0].set_title('Correlation Matrix')
df.groupby('date')['value'].mean().plot(ax=axes[1, 1], title='Trend Over Time')
plt.tight_layout()
plt.savefig('eda_report.png', dpi=150)
print("Chart saved to eda_report.png")
```

If chart rendering fails, fall back to a text summary table.

### Step 5: Report findings

Print a plain-language summary covering:
- Dataset shape and completeness (rows, columns, missing %)
- Key statistics (means, medians, notable distributions)
- Outliers or data quality issues found
- Patterns observed (correlations, group differences, trends)

## Error recovery

| Problem | Action |
|---------|--------|
| File not found | List directory contents with `os.listdir()`, ask user to confirm filename |
| Encoding error | Retry with `encoding='latin-1'` then `'cp1252'` |
| Mixed dtypes in column | Use `pd.to_numeric(col, errors='coerce')`, report coerced rows |
| Empty dataframe after filter | Warn user, show `value_counts()` for the filter column |
| Large dataset (>1M rows) | Use `df.sample(n=10000)` for profiling, full data for aggregation |
