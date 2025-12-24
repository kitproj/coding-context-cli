---
name: data-analysis
description: Analyze datasets, generate charts, and create summary reports. Use when the user needs to work with CSV, Excel, or other tabular data formats for analysis or visualization.
---

# Data Analysis

## When to use this skill
Use this skill when the user needs to:
- Analyze CSV or Excel files
- Generate charts and visualizations
- Calculate statistics and summaries
- Clean and transform data

## How to analyze data
1. Use pandas for data analysis:
   ```python
   import pandas as pd
   df = pd.read_csv('data.csv')
   summary = df.describe()
   ```

## How to create visualizations
1. Use matplotlib or seaborn for charts:
   ```python
   import matplotlib.pyplot as plt
   df.plot(kind='bar')
   plt.savefig('chart.png')
   ```
