# Machine Learning Workflow for "Will They Pay?"

You need to train the model first. Think of training as teaching the model by example: you feed it historical invoices where you **already know the outcome** (whether the patient paid within 30 days or not). The model studies these examples to find patterns. Once it learns those patterns, you can give it _new, unpaid_ invoices, and it will output a risk score.

A streamlined, step-by-step roadmap is detailed below to take your 5,000 rows of data and turn it into a trained, working model in Google Colab for your hackathon.

---

## The Machine Learning Workflow

1. **Prep Your Data** -> 2. **Define Target & Features** -> 3. **Split Data** -> 4. **Train Model** -> 5. **Evaluate Performance** -> 6. **Batch Score Invoices**

### Step 1: Prep Your Data (The Join)

Before touching any ML code, assemble your dataset into a single flat table (a Pandas DataFrame).

- Query your 5,000 historical invoices.
- Use the patient's ZIP code to map and merge the census data (**mean income** and **household size**).
- Handle any missing values (e.g., fill missing historical payment counts with `0`).

### Step 2: Define the Target ($y$) and Features ($X$)

You need to explicitly tell the model what to predict (the target) and what information to use to make that prediction (the features).

- **The Target ($y$):** Create a binary column. Look at historical invoices. If `days_to_payment <= 30`, set it to `1` (Paid on time). If it took longer or is still unpaid past 30 days, set it to `0` (Risk/Late).
- **The Features ($X$):** Drop identifiers like `patient_name` or `invoice_id` (the model can't learn patterns from random unique IDs). Keep your engineered features: `balance_amount`, `mean_income`, `household_size`, `historical_on_time_rate`, etc.

### Step 3: Split the Data (Train vs. Test)

Never evaluate your model on the same data it learned from—that's like giving a student the exact answers to the exam beforehand. Split your 5,000 rows into two groups:

- **Train Set (80% / 4,000 rows):** Used to train the model.
- **Test Set (20% / 1,000 rows):** Kept hidden until the end to test how accurate the model actually is.

### Step 4: Train the Model (The Code)

In Google Colab, this takes less than 10 lines of Python code using a library like `scikit-learn` or `xgboost`. Because your PRD mentions trying multiple baselines, starting with a **Random Forest** is an excellent choice—it requires almost no data scaling and handles mixed data types beautifully.

```python
from sklearn.model_selection import train_test_split
from sklearn.ensemble import RandomForestClassifier
from sklearn.metrics import classification_report, roc_auc_score

# 1. Separate features (X) and target (y)
X = df[['balance_amount', 'mean_income', 'household_size', 'historical_on_time_rate']]
y = df['paid_within_30_days']

# 2. Split into Train (80%) and Test (20%)
X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.2, random_state=42)

# 3. Initialize the model
model = RandomForestClassifier(n_estimators=100, random_state=42)

# 4. TRAIN the model (this is where the learning happens!)
model.fit(X_train, y_train)

print("Model training complete!")
```

### Step 5: Evaluate the Performance

Once `model.fit()` finishes running, evaluate its performance against your hidden Test set (`X_test`). Per your PRD, focus heavily on **Precision** for the "Unlikely to Pay" class.

```python
# Make predictions on the test set
predictions = model.predict(X_test)

# Print a report showing Precision, Recall, and F1-Score
print(classification_report(y_test, predictions))
```

### Step 6: Batch Score Your "Open" Invoices (The Hackathon Output)

Once you are happy with your model's precision, pull a completely separate dataset of **currently open, unpaid invoices** that the clinic is waiting on right now.

Pass these open invoices into your trained model using `predict_proba()`. This will output a probability score between 0% and 100% for each invoice.

```python
# Load your currently active open balances
open_invoices = pd.read_csv("open_balances.csv")

# Extract the exact same features used in training
X_open = open_invoices[['balance_amount', 'mean_income', 'household_size', 'historical_on_time_rate']]

# Get the probability of paying on time (column 1)
open_invoices['payment_probability'] = model.predict_proba(X_open)[:, 1]

# Map that probability to a risk tier for your UI
open_invoices['risk_tier'] = pd.cut(open_invoices['payment_probability'],
                                    bins=[0, 0.4, 0.7, 1.0],
                                    labels=['High Risk', 'Medium Risk', 'Low Risk'])

# Save this out as a CSV to feed your UI demo!
open_invoices.to_csv("scored_patient_invoices.csv", index=False)
```

Exporting this `scored_patient_invoices.csv` provides exactly what is needed to back your frontend demo, allowing your UI components to seamlessly display high-risk indicators and copy-paste messaging recommendations.
