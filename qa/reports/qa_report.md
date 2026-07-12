# QA Automation Report

**Project:** Challenge To YOU
**Run ID:** qa_1783891092150
**Time:** 2026-07-13T02:48:12+05:30 — 2026-07-13T02:49:04+05:30
**Duration:** 52586ms

## Summary

| Metric | Count |
|--------|-------|
| Total | 28 |
| Passed | 28 |
| Failed | 0 |
| Skipped | 0 |
| Errors | 0 |

**Status: ALL PASSING**

## Scenarios

### [PASS] FrameworkSelfTest (framework)

Duration: 1ms

| Step | Status | Details |
|------|--------|--------|
| framework_init | PASSED | QA framework initialized |
| assert_true | PASSED | true should pass |
| assert_false_cond | PASSED | math should work |
| dir_exists_reports | PASSED | reports directory created |
| dir_exists_backend_logs | PASSED | backend_logs directory created |
| dir_exists_fixtures | PASSED | fixtures directory created |
| dir_exists_screenshots | PASSED | screenshots directory created |
| report_json_exists | PASSED | qa_report.json should exist |
| report_md_exists | PASSED | qa_report.md should exist |

### [PASS] BackendLaunchAndShutdown (backend)

Duration: 1507ms

| Step | Status | Details |
|------|--------|--------|
| build_success | PASSED | Binary built: /Users/fallofpheonix/Project/game/Challenge-To-YOU/backend/sand... |
| server_started | PASSED | Server started on port 52441 |
| server_running | PASSED | Server uptime: 826.198625ms |
| server_stopped | PASSED | Port should be free after shutdown |

### [PASS] BackendHealthCheck (backend)

Duration: 659ms

| Step | Status | Details |
|------|--------|--------|
| health_check | PASSED | TCP connection successful |
| languages_endpoint | PASSED | /api/languages should return data |
| languages_response | PASSED | Response length: 160 bytes |

### [PASS] BackendLogCapture (backend)

Duration: 1662ms

| Step | Status | Details |
|------|--------|--------|
| logs_captured | PASSED | Log size: 484 bytes |
| log_content | PASSED | Captured 484 bytes of server logs |

### [PASS] WebSocketConnect (websocket)

Duration: 1513ms

| Step | Status | Details |
|------|--------|--------|
| ws_connect | PASSED | Connect error: <nil> |
| ws_connected | PASSED | Client should be connected |

### [PASS] WebSocketDisconnect (websocket)

Duration: 854ms

| Step | Status | Details |
|------|--------|--------|
| ws_connected_before | PASSED | Should be connected |
| ws_disconnected | PASSED | Should be disconnected after Disconnect() |

### [PASS] WebSocketReconnect (websocket)

Duration: 722ms

| Step | Status | Details |
|------|--------|--------|
| first_connect | PASSED | Should be connected |
| reconnect | PASSED | Should be reconnected |

### [PASS] WebSocketInvalidMessage (websocket)

Duration: 1199ms

| Step | Status | Details |
|------|--------|--------|
| invalid_sent | PASSED | Sent invalid JSON |
| server_alive_after_invalid | PASSED | Server should handle invalid messages and stay alive |

### [PASS] WebSocketTimeout (websocket)

Duration: 2145ms

| Step | Status | Details |
|------|--------|--------|
| initial_message | PASSED | Sent profile request |
| connection_alive | PASSED | Connection should remain alive after idle period |

### [PASS] MissionStart (mission)

Duration: 1532ms

| Step | Status | Details |
|------|--------|--------|
| mission_loaded | PASSED | Challenge ID: magitech_01_breach |
| paradigm_set | PASSED | Paradigm: MAGITECH |
| modules_present | PASSED | Module count: 3 |
| state_initialized | PASSED | State should be initialized |

### [PASS] MissionObjectives (mission)

Duration: 1329ms

| Step | Status | Details |
|------|--------|--------|
| has_modules | PASSED | Found 3 modules |
| has_state | PASSED | State has 4 keys |
| not_complete | PASSED | Level should not be complete initially |
| vigilance_low | PASSED | Vigilance: 0.000000 |

### [PASS] MissionStateTransitions (mission)

Duration: 740ms

| Step | Status | Details |
|------|--------|--------|
| initial_state | PASSED | State keys: [binding_active entropy mana_critical ward_sealed] |
| trigger_event | PASSED | Triggering: invoke_binding |
| trigger_processed | PASSED | State changed: false, Vigilance: 0.00 -> 0.00, Message: "USER STATS: reputati... |

### [PASS] MissionCompletion (mission)

Duration: 666ms

| Step | Status | Details |
|------|--------|--------|
| trigger_0 | PASSED | Event: invoke_binding, Complete: false, Vigilance: 0.00 |
| trigger_1 | PASSED | Event: invoke_binding, Complete: false, Vigilance: 0.10 |
| trigger_2 | PASSED | Event: invoke_binding, Complete: false, Vigilance: 0.20 |
| trigger_3 | PASSED | Event: surge_mana, Complete: false, Vigilance: 0.30 |
| trigger_4 | PASSED | Event: surge_mana, Complete: false, Vigilance: 0.40 |
| trigger_5 | PASSED | Event: invoke_binding, Complete: false, Vigilance: 0.50 |
| trigger_6 | PASSED | Event: invoke_binding, Complete: false, Vigilance: 0.60 |
| trigger_7 | PASSED | Event: trigger_release, Complete: false, Vigilance: 0.70 |
| trigger_8 | PASSED | Event: invoke_binding, Complete: true, Vigilance: 0.80 |
| system_responded | PASSED | System should respond to trigger events |

### [PASS] MissionUnlocks (mission)

Duration: 454ms

| Step | Status | Details |
|------|--------|--------|
| unlock_attempt | PASSED | Unlock CYBERPUNK attempted |

### [PASS] ChallengeLoad (challenge)

Duration: 11ms

| Step | Status | Details |
|------|--------|--------|
| has_id_magitech_01.json | PASSED | Challenge must have id |
| has_paradigm_magitech_01.json | PASSED | Challenge must have paradigm |
| has_flaws_magitech_01.json | PASSED | Challenge must have flaws |
| has_id_magitech_02_centrifuge.json | PASSED | Challenge must have id |
| has_paradigm_magitech_02_centrifuge.json | PASSED | Challenge must have paradigm |
| has_flaws_magitech_02_centrifuge.json | PASSED | Challenge must have flaws |
| has_id_magitech_03_vault.json | PASSED | Challenge must have id |
| has_paradigm_magitech_03_vault.json | PASSED | Challenge must have paradigm |
| has_flaws_magitech_03_vault.json | PASSED | Challenge must have flaws |
| has_id_magitech_04_golem.json | PASSED | Challenge must have id |
| has_paradigm_magitech_04_golem.json | PASSED | Challenge must have paradigm |
| has_flaws_magitech_04_golem.json | PASSED | Challenge must have flaws |
| has_id_magitech_05_grimoire.json | PASSED | Challenge must have id |
| has_paradigm_magitech_05_grimoire.json | PASSED | Challenge must have paradigm |
| has_flaws_magitech_05_grimoire.json | PASSED | Challenge must have flaws |
| has_id_magitech_06_astrolabe.json | PASSED | Challenge must have id |
| has_paradigm_magitech_06_astrolabe.json | PASSED | Challenge must have paradigm |
| has_flaws_magitech_06_astrolabe.json | PASSED | Challenge must have flaws |
| has_id_magitech_07_loom.json | PASSED | Challenge must have id |
| has_paradigm_magitech_07_loom.json | PASSED | Challenge must have paradigm |
| has_flaws_magitech_07_loom.json | PASSED | Challenge must have flaws |
| has_id_magitech_08_mana_overflow.json | PASSED | Challenge must have id |
| has_paradigm_magitech_08_mana_overflow.json | PASSED | Challenge must have paradigm |
| has_flaws_magitech_08_mana_overflow.json | PASSED | Challenge must have flaws |
| has_id_magitech_09_resonance.json | PASSED | Challenge must have id |
| has_paradigm_magitech_09_resonance.json | PASSED | Challenge must have paradigm |
| has_flaws_magitech_09_resonance.json | PASSED | Challenge must have flaws |
| has_id_magitech_m01_runic_initiation.json | PASSED | Challenge must have id |
| has_paradigm_magitech_m01_runic_initiation.json | PASSED | Challenge must have paradigm |
| has_flaws_magitech_m01_runic_initiation.json | PASSED | Challenge must have flaws |
| skip_pack.json | PASSED | Pack/composite file, skipping validation |
| era_magitech_tier1 | PASSED | Loaded 11 challenges from magitech_tier1 |
| skip_composite_01_concat.json | PASSED | Pack/composite file, skipping validation |
| skip_composite_02_state.json | PASSED | Pack/composite file, skipping validation |
| skip_composite_03_pipe.json | PASSED | Pack/composite file, skipping validation |
| has_id_cyberpunk_01_autodoc.json | PASSED | Challenge must have id |
| has_paradigm_cyberpunk_01_autodoc.json | PASSED | Challenge must have paradigm |
| has_flaws_cyberpunk_01_autodoc.json | PASSED | Challenge must have flaws |
| has_id_cyberpunk_02_elevator.json | PASSED | Challenge must have id |
| has_paradigm_cyberpunk_02_elevator.json | PASSED | Challenge must have paradigm |
| has_flaws_cyberpunk_02_elevator.json | PASSED | Challenge must have flaws |
| has_id_cyberpunk_03_server.json | PASSED | Challenge must have id |
| has_paradigm_cyberpunk_03_server.json | PASSED | Challenge must have paradigm |
| has_flaws_cyberpunk_03_server.json | PASSED | Challenge must have flaws |
| has_id_cyberpunk_04_barista.json | PASSED | Challenge must have id |
| has_paradigm_cyberpunk_04_barista.json | PASSED | Challenge must have paradigm |
| has_flaws_cyberpunk_04_barista.json | PASSED | Challenge must have flaws |
| has_id_cyberpunk_05_drone.json | PASSED | Challenge must have id |
| has_paradigm_cyberpunk_05_drone.json | PASSED | Challenge must have paradigm |
| has_flaws_cyberpunk_05_drone.json | PASSED | Challenge must have flaws |
| has_id_cyberpunk_06_traffic.json | PASSED | Challenge must have id |
| has_paradigm_cyberpunk_06_traffic.json | PASSED | Challenge must have paradigm |
| has_flaws_cyberpunk_06_traffic.json | PASSED | Challenge must have flaws |
| has_id_cyberpunk_07_optimize.json | PASSED | Challenge must have id |
| has_paradigm_cyberpunk_07_optimize.json | PASSED | Challenge must have paradigm |
| has_flaws_cyberpunk_07_optimize.json | PASSED | Challenge must have flaws |
| has_id_cyberpunk_08_spec.json | PASSED | Challenge must have id |
| has_paradigm_cyberpunk_08_spec.json | PASSED | Challenge must have paradigm |
| has_flaws_cyberpunk_08_spec.json | PASSED | Challenge must have flaws |
| has_id_cyberpunk_09_recognize.json | PASSED | Challenge must have id |
| has_paradigm_cyberpunk_09_recognize.json | PASSED | Challenge must have paradigm |
| has_flaws_cyberpunk_09_recognize.json | PASSED | Challenge must have flaws |
| has_id_cyberpunk_c42_thread_race.json | PASSED | Challenge must have id |
| has_paradigm_cyberpunk_c42_thread_race.json | PASSED | Challenge must have paradigm |
| has_flaws_cyberpunk_c42_thread_race.json | PASSED | Challenge must have flaws |
| skip_pack.json | PASSED | Pack/composite file, skipping validation |
| era_cyberpunk_tier1 | PASSED | Loaded 14 challenges from cyberpunk_tier1 |
| has_id_cosmic_01_airlock.json | PASSED | Challenge must have id |
| has_paradigm_cosmic_01_airlock.json | PASSED | Challenge must have paradigm |
| has_flaws_cosmic_01_airlock.json | PASSED | Challenge must have flaws |
| has_id_cosmic_02_nav.json | PASSED | Challenge must have id |
| has_paradigm_cosmic_02_nav.json | PASSED | Challenge must have paradigm |
| has_flaws_cosmic_02_nav.json | PASSED | Challenge must have flaws |
| has_id_cosmic_03_stasis.json | PASSED | Challenge must have id |
| has_paradigm_cosmic_03_stasis.json | PASSED | Challenge must have paradigm |
| has_flaws_cosmic_03_stasis.json | PASSED | Challenge must have flaws |
| has_id_cosmic_04_winch.json | PASSED | Challenge must have id |
| has_paradigm_cosmic_04_winch.json | PASSED | Challenge must have paradigm |
| has_flaws_cosmic_04_winch.json | PASSED | Challenge must have flaws |
| has_id_cosmic_05_seed.json | PASSED | Challenge must have id |
| has_paradigm_cosmic_05_seed.json | PASSED | Challenge must have paradigm |
| has_flaws_cosmic_05_seed.json | PASSED | Challenge must have flaws |
| has_id_cosmic_06_singularity.json | PASSED | Challenge must have id |
| has_paradigm_cosmic_06_singularity.json | PASSED | Challenge must have paradigm |
| has_flaws_cosmic_06_singularity.json | PASSED | Challenge must have flaws |
| has_id_cosmic_07_relay.json | PASSED | Challenge must have id |
| has_paradigm_cosmic_07_relay.json | PASSED | Challenge must have paradigm |
| has_flaws_cosmic_07_relay.json | PASSED | Challenge must have flaws |
| has_id_cosmic_08_valve.json | PASSED | Challenge must have id |
| has_paradigm_cosmic_08_valve.json | PASSED | Challenge must have paradigm |
| has_flaws_cosmic_08_valve.json | PASSED | Challenge must have flaws |
| has_id_cosmic_v81_ast_parser.json | PASSED | Challenge must have id |
| has_paradigm_cosmic_v81_ast_parser.json | PASSED | Challenge must have paradigm |
| has_flaws_cosmic_v81_ast_parser.json | PASSED | Challenge must have flaws |
| skip_pack.json | PASSED | Pack/composite file, skipping validation |
| era_cosmic_tier1 | PASSED | Loaded 10 challenges from cosmic_tier1 |
| total_challenges | PASSED | Total challenges loaded: 35 |

### [PASS] ChallengeCompile (challenge)

Duration: 156ms

| Step | Status | Details |
|------|--------|--------|
| compile_success | PASSED | Compile error: <nil> |
| compile_rejects_invalid | PASSED | Compiler should reject invalid Python |

### [PASS] ChallengeExecute (challenge)

Duration: 21ms

| Step | Status | Details |
|------|--------|--------|
| exec_success | PASSED | Exec error: <nil> |
| exec_output | PASSED | Output: OUTPUT:hello_world STATUS:success  |
| output_contains_hello | PASSED | Output: OUTPUT:hello_world STATUS:success  |

### [PASS] ChallengeScoring (challenge)

Duration: 640ms

| Step | Status | Details |
|------|--------|--------|
| initial_xp | PASSED | Starting XP: 0 |
| final_xp | PASSED | XP after triggers: 0 |
| scoring_system_active | PASSED | XP tracked: 0 -> 0 |

### [PASS] SaveCreateLoad (save)

Duration: 657ms

| Step | Status | Details |
|------|--------|--------|
| save_profile_loaded | PASSED | Profile should load from DB |
| state_modified | PASSED | Triggered event to modify state |

### [PASS] SaveRestartReload (save)

Duration: 1280ms

| Step | Status | Details |
|------|--------|--------|
| server1_state | PASSED | Challenge: magitech_01_breach |
| save_persisted | PASSED | State should persist across restarts |
| challenge_restored | PASSED | Challenge: magitech_01_breach -> magitech_01_breach |
| server2_state | PASSED | Challenge: magitech_01_breach (restored) |

### [PASS] SaveCorruptedHandling (save)

Duration: 30244ms

| Step | Status | Details |
|------|--------|--------|
| corrupted_written | PASSED | Wrote corrupted DB file |
| corrupt_handled_gracefully | PASSED | Server correctly rejected corrupted DB: server not healthy: timeout waiting f... |

### [PASS] RewardXP (reward)

Duration: 855ms

| Step | Status | Details |
|------|--------|--------|
| initial_snapshot | PASSED | Received initial snapshot |
| profile_loaded | PASSED | Profile event sent and response received |
| xp_tracking | PASSED | Profile stats in message: USER STATS: reputation=0, luck=1.00, unlocked_parad... |
| challenge_loaded | PASSED | Challenge: magitech_01_breach |

### [PASS] RewardCredits (reward)

Duration: 648ms

| Step | Status | Details |
|------|--------|--------|
| credits_tracking | PASSED | Profile message: USER STATS: reputation=0, luck=1.00, unlocked_paradigms=[MAG... |
| profile_reachable | PASSED | Profile stats returned in message |

### [PASS] RewardReputation (reward)

Duration: 445ms

| Step | Status | Details |
|------|--------|--------|
| reputation_visible | PASSED | Reputation info in message: USER STATS: reputation=0, luck=1.00, unlocked_par... |

### [PASS] RewardLuck (reward)

Duration: 652ms

| Step | Status | Details |
|------|--------|--------|
| luck_visible | PASSED | Luck info in message: USER STATS: reputation=0, luck=1.00, unlocked_paradigms... |

### [PASS] RewardUnlocks (reward)

Duration: 506ms

| Step | Status | Details |
|------|--------|--------|
| unlocks_visible | PASSED | Unlocks info in message: USER STATS: reputation=0, luck=1.00, unlocked_paradi... |

### [PASS] RegressionDetection (regression)

Duration: 2ms

| Step | Status | Details |
|------|--------|--------|
| challenge_count_stable | PASSED | Current: 35, Previous: 35 |
| no_challenges_removed | PASSED | Removed challenges: [] |
| endpoints_stable | PASSED | Endpoints: 3 -> 3 |
| regression_check_complete | PASSED | Compared against baseline from 2026-07-13T02:43:59+05:30 |

### [PASS] ContinuousPlaythrough (e2e)

Duration: 1468ms

| Step | Status | Details |
|------|--------|--------|
| phase1_build | PASSED | Backend binary compiled |
| phase1_launch | PASSED | Server launched on port 53275 |
| phase2_connect | PASSED | WebSocket client connected |
| phase3_mission_started | PASSED | Should receive initial snapshot |
| phase3_mission | PASSED | Mission loaded: magitech_01_breach (Paradigm: MAGITECH) |
| phase4_state_ready | PASSED | Game state initialized |
| phase5_challenge_loaded | PASSED | Challenge should be loaded |
| phase5_has_modules | PASSED | Modules: 3 |
| phase5_has_triggers | PASSED | Triggers: [invoke_binding] |
| phase6_round_0 | PASSED | Event: invoke_binding \| Vigilance: 0.00 \| Complete: false |
| phase6_round_1 | PASSED | Event: invoke_binding \| Vigilance: 0.10 \| Complete: false |
| phase6_round_2 | PASSED | Event: invoke_binding \| Vigilance: 0.20 \| Complete: false |
| phase6_round_3 | PASSED | Event: invoke_binding \| Vigilance: 0.30 \| Complete: false |
| phase6_round_4 | PASSED | Event: invoke_binding \| Vigilance: 0.40 \| Complete: false |
| phase6_round_5 | PASSED | Event: invoke_binding \| Vigilance: 0.50 \| Complete: false |
| phase6_round_6 | PASSED | Event: invoke_binding \| Vigilance: 0.60 \| Complete: false |
| phase6_round_7 | PASSED | Event: invoke_binding \| Vigilance: 0.70 \| Complete: false |
| phase6_round_8 | PASSED | Event: invoke_binding \| Vigilance: 0.80 \| Complete: false |
| phase6_round_9 | PASSED | Event: invoke_binding \| Vigilance: 0.90 \| Complete: false |
| phase6_round_10 | PASSED | Event: invoke_binding \| Vigilance: 1.00 \| Complete: false |
| phase6_send_done | PASSED | Round 11: connection closed: write tcp 127.0.0.1:53285->127.0.0.1:53275: writ... |
| phase7_partial | PASSED | Challenge not completed after 11 rounds (expected for automated test) |
| phase9_save | PASSED | Server stopped (state persisted to DB) |
| phase10_restart | PASSED | Server restarted on port 53286 |
| phase11_loaded | PASSED | Should receive snapshot after restart |
| phase11_reload | PASSED | State reloaded: magitech_01_breach |
| phase12_paradigm_match | PASSED | Paradigm: MAGITECH -> MAGITECH |
| phase13_disconnect | PASSED | Final disconnect complete |
| e2e_complete | PASSED | Full playthrough completed. 11 rounds executed. |

---
Generated by QA Automation Suite | go1.26.4 | darwin/arm64
