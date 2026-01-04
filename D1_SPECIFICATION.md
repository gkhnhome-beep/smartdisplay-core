# DESIGN PHASE D1 SPECIFICATION

**Phase:** DESIGN Phase D1  
**Goal:** Define first-boot texts, tone, and microcopy  
**Date:** January 4, 2026  
**Status:** IN PROGRESS - DEFINING

---

## Product Principles Foundation

All text must align with SmartDisplay's core principles:

1. **Calm** - Soothing, never urgent or pushy
2. **Predictable** - Clear expectations, no surprises
3. **Respectful** - Values user time and intelligence
4. **Protective** - Security-conscious, not paranoid

---

## First-Boot Step Definitions

### STEP 1: Welcome

**Purpose:** Greet user and establish calm tone

**i18n Keys & Text:**

| Key | Context | English Source |
|-----|---------|-----------------|
| `firstboot.welcome.heading` | Main heading | Welcome to SmartDisplay |
| `firstboot.welcome.subtitle` | Subheading (reduced_motion variant) | Let's get you started. |
| `firstboot.welcome.subtitle.full` | Subheading (standard variant) | Let's set up your smart alarm panel in just a few moments. |
| `firstboot.welcome.description` | Description line (standard) | SmartDisplay protects your home with calm, intelligent alarm management. |
| `firstboot.welcome.description.short` | Description line (reduced_motion) | A smart alarm panel for your home. |
| `firstboot.welcome.tone` | Internal: tone descriptor | Welcoming, professional, reassuring |

**Voice Variant:**
```
"Welcome to SmartDisplay. A smart alarm panel for your home."
```

**Accessibility Variants:**

- **Reduced Motion:** "Let's get you started."
- **High Contrast:** Bold heading, simple font
- **Large Text:** Simpler phrasing, fewer words

---

### STEP 2: Language Confirmation

**Purpose:** Confirm language preference, allow change

**i18n Keys & Text:**

| Key | Context | English Source |
|-----|---------|-----------------|
| `firstboot.language.heading` | Main heading | Language |
| `firstboot.language.current` | Display current (standard) | Your language is set to English. |
| `firstboot.language.current.short` | Display current (reduced_motion) | Language: English |
| `firstboot.language.question` | Question (standard) | Would you like to change it? |
| `firstboot.language.question.short` | Question (reduced_motion) | (Change in next step if needed.) |
| `firstboot.language.option_en` | English option label | English |
| `firstboot.language.option_tr` | Turkish option label | Türkçe |
| `firstboot.language.tone` | Internal: tone descriptor | Clear, simple, informative |

**Voice Variant:**
```
"Your language is set to English. You can change it if you'd like."
```

**Accessibility Variants:**

- **Reduced Motion:** "Language: English"
- **High Contrast:** Clear radio button options
- **Large Text:** Larger option labels, more spacing

---

### STEP 3: Home Assistant Status Check

**Purpose:** Show HA connection status (information only, no setup)

**i18n Keys & Text:**

| Key | Context | English Source |
|-----|---------|-----------------|
| `firstboot.ha_check.heading` | Main heading | Home Assistant |
| `firstboot.ha_check.description` | Description (standard) | SmartDisplay integrates with Home Assistant for device control. |
| `firstboot.ha_check.description.short` | Description (reduced_motion) | Connection with Home Assistant |
| `firstboot.ha_check.status_connected` | Status when connected | Connected ✓ |
| `firstboot.ha_check.status_disconnected` | Status when disconnected | Not connected yet |
| `firstboot.ha_check.note_connected` | Note when connected (standard) | Your Home Assistant is ready. Setup will continue. |
| `firstboot.ha_check.note_connected.short` | Note when connected (reduced_motion) | Home Assistant is ready. |
| `firstboot.ha_check.note_disconnected` | Note when disconnected (standard) | SmartDisplay works without Home Assistant, but works best with it. You can add it later. |
| `firstboot.ha_check.note_disconnected.short` | Note when disconnected (reduced_motion) | You can connect Home Assistant later. |
| `firstboot.ha_check.tone` | Internal: tone descriptor | Informative, reassuring, non-blocking |

**Voice Variant (Connected):**
```
"Home Assistant is connected. You're all set to continue."
```

**Voice Variant (Disconnected):**
```
"Home Assistant is not connected yet. You can add it later if you'd like."
```

**Accessibility Variants:**

- **Reduced Motion:** Simpler status display, no progress animations
- **High Contrast:** Clear status indicator (green/gray)
- **Large Text:** Larger status font, simpler note text

---

### STEP 4: Alarm Role Explanation

**Purpose:** Explain system's primary purpose (information only)

**i18n Keys & Text:**

| Key | Context | English Source |
|-----|---------|-----------------|
| `firstboot.alarm_role.heading` | Main heading | What SmartDisplay Does |
| `firstboot.alarm_role.purpose` | Main explanation (standard) | SmartDisplay is your personal alarm panel. It learns your patterns and protects your home with intelligent, calm monitoring. |
| `firstboot.alarm_role.purpose.short` | Main explanation (reduced_motion) | SmartDisplay is an intelligent alarm panel for your home. |
| `firstboot.alarm_role.feature_learn` | Feature 1 (standard) | Learns your daily patterns to reduce false alarms. |
| `firstboot.alarm_role.feature_learn.short` | Feature 1 (reduced_motion) | Learns your patterns. |
| `firstboot.alarm_role.feature_calm` | Feature 2 (standard) | Notifies you calmly, never panics. |
| `firstboot.alarm_role.feature_calm.short` | Feature 2 (reduced_motion) | Never panics. |
| `firstboot.alarm_role.feature_guest` | Feature 3 (standard) | Manages guest access safely and smartly. |
| `firstboot.alarm_role.feature_guest.short` | Feature 3 (reduced_motion) | Manages guest access. |
| `firstboot.alarm_role.tone` | Internal: tone descriptor | Confident, reassuring, focused |

**Voice Variant:**
```
"SmartDisplay is an intelligent alarm panel that learns your home. It protects you calmly and never panics."
```

**Accessibility Variants:**

- **Reduced Motion:** Single sentence, bullet points removed
- **High Contrast:** Bold feature headers
- **Large Text:** One feature per screen option, larger fonts

---

### STEP 5: Ready / Completion

**Purpose:** Confirm setup complete, begin normal operation

**i18n Keys & Text:**

| Key | Context | English Source |
|-----|---------|-----------------|
| `firstboot.ready.heading` | Main heading | You're All Set |
| `firstboot.ready.message` | Main message (standard) | SmartDisplay is ready to protect your home. Start by exploring the main panel. |
| `firstboot.ready.message.short` | Main message (reduced_motion) | SmartDisplay is ready. |
| `firstboot.ready.cta` | Call-to-action button | Start |
| `firstboot.ready.tone` | Internal: tone descriptor | Confident, welcoming, ready to go |

**Voice Variant:**
```
"You're all set. SmartDisplay is ready to protect your home."
```

**Accessibility Variants:**

- **Reduced Motion:** Minimal text, focus on CTA button
- **High Contrast:** Large button with clear text
- **Large Text:** Larger button, bigger font

---

## System Messages

Messages displayed in various system states:

### During Setup
```
system_message: "Setup in progress"
subtitle: "Step {current} of {total}"
context: Shown on UI endpoints during firstboot mode
```

**i18n Keys:**
```
firstboot.system.setup_in_progress = "Setup in progress"
firstboot.system.step_counter = "Step {current} of {total}"
firstboot.system.step_counter.short = "Step {current}"
```

### Nearly Complete
```
message: "Almost ready"
subtitle: "One more step"
context: Shown at step 4 (Alarm Role Explanation)
```

**i18n Keys:**
```
firstboot.system.almost_ready = "Almost ready"
firstboot.system.one_more_step = "One more step"
```

### Upon Completion
```
message: "Setup complete"
subtitle: "Welcome to SmartDisplay"
context: Shown after POST /api/setup/firstboot/complete succeeds
```

**i18n Keys:**
```
firstboot.system.setup_complete = "Setup complete"
firstboot.system.welcome_aboard = "Welcome to SmartDisplay"
```

---

## Tone Guidelines

### DO
✅ Use active voice ("SmartDisplay protects")  
✅ Be specific ("intelligent alarm panel" not "smart system")  
✅ Use positive framing ("learns your patterns" not "monitors you")  
✅ Be honest about capabilities ("protects your home" not "provides complete security")  
✅ Respect user time (short, scannable text)  
✅ Use contractions ("you're" not "you are")  
✅ Address user directly ("your home", "you can")  

### DON'T
❌ Use tech jargon ("MQTT", "webhook", "REST API")  
❌ Use marketing speak ("revolutionary", "best-in-class")  
❌ Make promises ("100% safe", "never fails")  
❌ Be cute or try too hard ("Let's get this party started!")  
❌ Use fear or urgency ("Hackers could", "Act now")  
❌ Apologize unnecessarily ("Sorry, we need to ask...")  

---

## Localization Strategy

### English Source
- Written in clear, simple English
- Standard variant as primary
- Reduced-motion variant as shorter alternative
- Voice variants conversational but clear

### Turkish Localization

**Language Principles:**
- Maintain calm, professional tone
- Avoid slang or colloquialisms
- Use formal "siz" unless specified otherwise
- Match English word counts (within 10%)

**Turkish Variants (To Be Completed):**

| English Key | Turkish Key | Turkish Source | Notes |
|-------------|------------|-----------------|-------|
| `firstboot.welcome.heading` | `firstboot.welcome.heading.tr` | Hoş geldiniz SmartDisplay'e | Warm, formal greeting |
| `firstboot.welcome.subtitle.full` | `firstboot.welcome.subtitle.full.tr` | Akıllı alarm panelini birkaç adımda kuralım. | Gentle guidance |
| `firstboot.language.option_tr` | `firstboot.language.option_tr.tr` | Türkçe | Self-referential |

---

## Accessibility Mapping

### For `reduced_motion` users:
- Use shorter variants (marked `.short`)
- Remove animation descriptions
- Simpler sentence structure
- Direct statements only

### For `high_contrast` users:
- Ensure text has sufficient contrast
- No decorative text needed
- Clear headers and labels
- Simple font recommendations

### For `large_text` users:
- Shorter line lengths
- Simpler vocabulary
- Fewer clauses per sentence
- More white space between ideas

**Implementation Note:** These are UI/rendering concerns. The i18n keys provide the text; the UI layer applies styling based on `high_contrast` and `large_text` preferences.

---

## Voice Integration

For future voice feedback (FAZ 81), these variants provide clearer cadence:

### Rule for Voice Text:
- No punctuation in the voice variant (periods implied by natural cadence)
- 5-15 words typically optimal
- Active voice preferred
- No abbreviations or symbols

**Examples:**

English Text: "SmartDisplay protects your home with calm, intelligent alarm management."  
Voice Variant: "SmartDisplay is an intelligent alarm panel that protects your home"

English Text: "You can connect Home Assistant later if you'd like."  
Voice Variant: "You can add Home Assistant later if you would like"

---

## Consistency Checklist

Each step's copy should:

✅ Fit Product Principles (Calm, Predictable, Respectful, Protective)  
✅ Use simple language (8th-grade reading level max)  
✅ Respect user time (max 2 short sentences)  
✅ Avoid jargon and marketing speak  
✅ Have professional, warm tone  
✅ Include reduced_motion variant  
✅ Include voice variant  
✅ Align with similar steps' style  
✅ Provide clear next action (when applicable)  

---

## Character Limits (for UI planning)

| Element | English | Turkish | Limit |
|---------|---------|---------|-------|
| Step heading | ~20 chars | ~25 chars | 30 |
| Subtitle | ~50 chars | ~60 chars | 80 |
| Description | ~100 chars | ~120 chars | 150 |
| Short variant | ~30 chars | ~40 chars | 50 |
| Voice variant | ~50 chars | ~60 chars | 80 |

---

## Next Implementation Steps (D1 Completion)

1. **Create i18n keys** in the localization system
2. **Add English source text** to i18n/en.json
3. **Create Turkish translations** for all keys
4. **Validate tone** against Product Principles
5. **Test accessibility** variants with real users (if available)
6. **Document voice** integration path
7. **Create reference** for UI developers

---

## Validation Against Product Principles

### Calm ✓
- No urgency language
- Reassuring tone throughout
- "Never panics" reassurance in alarm role
- Acknowledgment that setup takes "just a few moments"

### Predictable ✓
- Clear step progression (1-5)
- Explicit statement of what happens next
- No surprise requirements
- Transparent about Home Assistant (optional, can add later)

### Respectful ✓
- Acknowledges user intelligence (no patronizing)
- Offers options (language change)
- Explains WHY (not just what)
- Values time (concise text)

### Protective ✓
- Emphasizes safety ("protects your home")
- Calm monitoring approach
- Security-conscious (HA integration optional)
- No false promises

---

## Tone Examples

**GOOD (Calm, Professional):**
> "SmartDisplay protects your home with intelligent, calm monitoring."

**BAD (Too marketing):**
> "Experience the ultimate smart home revolution with SmartDisplay!"

**GOOD (Clear, Respectful):**
> "You can connect Home Assistant later if you'd like."

**BAD (Pushy):**
> "You really should set up Home Assistant now for full protection!"

**GOOD (Predictable):**
> "Your language is set to English."

**BAD (Vague):**
> "Language settings are being configured..."

---

## Sign-Off

This D1 Specification defines:

- ✅ Complete text for all 5 first-boot steps
- ✅ i18n key structure and English sources
- ✅ Accessibility variants for reduced_motion
- ✅ Voice-friendly text alternatives
- ✅ System messages and tone guidelines
- ✅ Turkish localization strategy
- ✅ Validation against Product Principles
- ✅ Character limits for UI planning

**Ready for implementation:** D1 can now move to integration phase where these keys are added to the i18n system and the FirstBootManager is enhanced to use them.

---

**Status:** SPECIFICATION COMPLETE - READY FOR IMPLEMENTATION
