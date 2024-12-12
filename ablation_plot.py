import pandas as pd
import matplotlib.pyplot as plt
import seaborn as sns
import os

# Define experiment IDs: 0 for baseline, 1-11 for varying proportions
experiment_ids = range(0, 12)

# Prepare a list to store survival metrics
survival_data = []

for exp_id in experiment_ids:
    base_path = f"visualization_output/csv_data/experiment_{exp_id}"
    agent_file = os.path.join(base_path, "agent_records.csv")
    
    # Load the agent records
    agents_df = pd.read_csv(agent_file)
    
    # Convert IsAlive to boolean if needed
    if agents_df['IsAlive'].dtype == object:
        agents_df['IsAlive'] = agents_df['IsAlive'].map(lambda x: True if str(x).lower() == 'true' else False)
    
    # Classify agents as good or evil based on SpecialNote
    # Assuming SpecialNote field exists and encodes alignment like 'E1_C1' for good, 'E3_C1' for evil.
    agents_df['Alignment'] = agents_df['SpecialNote'].apply(
        lambda note: 'Good' if isinstance(note, str) and note.startswith('E1') else
                     ('Evil' if isinstance(note, str) and note.startswith('E3') else 'Neutral')
    )
    
    # Group by AgentID and determine survival for each agent
    agent_survival = agents_df.groupby('AgentID').apply(
        lambda grp: grp[grp['IsAlive'] == True]['TurnNumber'].max() if any(grp['IsAlive']) else 0
    )

    # Merge survival times back to agent info to know their alignment
    # We take the alignment from the first record of that agent (assuming alignment doesn't change)
    agent_alignment = agents_df.groupby('AgentID')['Alignment'].first()

    # Create separate arrays for good and evil agents
    good_survivals = agent_survival[agent_alignment == 'Good']
    evil_survivals = agent_survival[agent_alignment == 'Evil']
    neutral_survivals = agent_survival[agent_alignment == 'Neutral']

    # Compute E3 proportion based on experiment ID:
    if exp_id == 0:
        e3_prop = None
    else:
        e3_prop = (exp_id - 1) * 10  # increments of 10% from experiment 1 to 11

    # Compute statistics for good agents
    if len(good_survivals) > 0:
        good_mean = good_survivals.mean()
        good_min = good_survivals.min()
        good_max = good_survivals.max()
    else:
        good_mean = good_min = good_max = None

    # Compute statistics for evil agents
    if len(evil_survivals) > 0:
        evil_mean = evil_survivals.mean()
        evil_min = evil_survivals.min()
        evil_max = evil_survivals.max()
    else:
        evil_mean = evil_min = evil_max = None

    # Compute statistics for neutral agents
    if len(neutral_survivals) > 0:
        neutral_mean = neutral_survivals.mean()
        neutral_min = neutral_survivals.min()
        neutral_max = neutral_survivals.max()
    else:
        neutral_mean = neutral_min = neutral_max = None

    survival_data.append({
        'ExperimentID': exp_id,
        'E3_Proportion_%': e3_prop,
        'Good_mean': good_mean,
        'Good_min': good_min,
        'Good_max': good_max,
        'Evil_mean': evil_mean,
        'Evil_min': evil_min,
        'Evil_max': evil_max,
        'Neutral_mean': neutral_mean,
        'Neutral_min': neutral_min,
        'Neutral_max': neutral_max
    })

# Create a DataFrame of the results
survival_df = pd.DataFrame(survival_data)

# Filter only the experiments with varying E3 proportion (1 to 11)
comparison_df = survival_df[survival_df['ExperimentID'] > 0]

sns.set(style="whitegrid")
plt.figure(figsize=(10, 6))

# We will plot good and evil agents separately.
# Use error bars to represent min and max range. We'll use the mean with vertical error bars.
# For good agents:
good_y = comparison_df['Good_mean']
good_x = comparison_df['E3_Proportion_%']
good_yerr = [good_y - comparison_df['Good_min'], comparison_df['Good_max'] - good_y]

plt.errorbar(good_x, good_y, yerr=good_yerr, fmt='-o', label='Good Agents', capsize=5, alpha=0.9, color='blue')

# For evil agents:
evil_y = comparison_df['Evil_mean']
evil_x = comparison_df['E3_Proportion_%']
evil_yerr = [evil_y - comparison_df['Evil_min'], comparison_df['Evil_max'] - evil_y]

plt.errorbar(evil_x, evil_y, yerr=evil_yerr, fmt='-o', label='Evil Agents', capsize=5, alpha=0.9, color='red')

plt.title('Survival Turns by Agent Alignment vs. Evilness Percentage (C=3, E1↔E3 Mix)')
plt.xlabel('Evilness (E3) Proportion (%)')
plt.ylabel('Survival Turns (Mean ± Range)')

# Add dashed line for baseline (experiment_0)
baseline_survival_df = survival_df[survival_df['ExperimentID'] == 0]
neutral_mean = baseline_survival_df['Neutral_mean'].values[0]
neutral_max = baseline_survival_df['Neutral_max'].values[0]
neutral_min = baseline_survival_df['Neutral_min'].values[0]
plt.axhline(y=neutral_mean, color='black', linestyle='--', alpha=0.5, label='Baseline: Neutral Agents')
plt.axhline(y=neutral_max, color='black', linestyle='--', alpha=0.5)
plt.axhline(y=neutral_min, color='black', linestyle='--', alpha=0.5)

plt.legend()
plt.tight_layout()
plt.show()