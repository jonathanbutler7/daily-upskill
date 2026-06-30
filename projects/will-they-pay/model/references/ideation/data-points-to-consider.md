# Data Points to consider for model

Here are some high-yield data points you can add to your feature engineering list, categorized by how they influence the model:

### **1\. Financial & Insurance Context (The "Ability" to Pay)**

* **Insurance Coverage Type:** Is the patient Self-Pay (uninsured), Commercially Insured, or on a government plan like Medicaid? Uninsured patients typically carry higher risk for 30-day defaults.  
* **Total Out-of-Pocket Balance:** The absolute dollar amount owed. A $50 co-pay has a much higher likelihood of being paid within 30 days than a $2,500 bill for a dental implant.  
* **Patient Responsibility Ratio:** The percentage of the total bill that the patient owes vs. what insurance covered.

### **2\. Historical & Behavioral Data (The "Willingness" to Pay)**

* **Patient Tenure:** How many months or years has this person been a patient at the practice? Established, loyal patients are statistically much more likely to settle accounts quickly than one-time emergency patients.  
* **Appointment Reliability (No-Show Rate):** The ratio of missed/canceled appointments to completed ones. Patients who respect the practice's time tend to respect the billing cycle.  
* **Historical Payment Speed:** If they are a returning patient, what is their historical median "Days to Pay"?

### **3\. Service & Clinical Context (The "Urgency" Factor)**

* **Procedure Type:** Was the visit for preventative care (routine cleaning) or an emergency/restorative procedure (root canal, crown)? Preventative care is planned and budgeted for; emergency care is often an unexpected financial shock.  
* **Future Appointments Scheduled:** Does the patient have another appointment booked in the next 30–60 days? If they need to come back for follow-up work, they are highly incentivized to clear their balance so they aren't turned away at the front desk.

### **💡 Hackathon Pro-Tips for Your Model**

* **The "Payday" Temporal Feature:** Extract the day of the month the bill was issued. Bills hitting a patient's inbox around the 1st or the 15th (standard paydays) often see a much faster turnaround than bills sent on the 22nd.  
* **The "Cold Start" Problem:** For brand-new patients, your "past history" features will be blank. Create a binary flag called `is_new_patient`. When this is `1`, your model will automatically learn to rely heavier on your demographic features (like the zipcode income proxy) and the procedure type.  
* **Target Variable Definition:** Make sure your target is strictly engineered. Instead of just "paid vs unpaid," make it a binary classification: `1` if `days_to_payment <= 30`, and `0` if `days_to_payment > 30` (including unpaid).

