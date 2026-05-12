# Bugs pendientes post-Phase 10

## 1. ParseStructArrayItemsAsPropertyCollections item splitting incompleto
- Impacto: ~14 conversaciones reply-only + cithub_arrival_ai_a_dlg
- Legacy: 21e/11r/2s | V2: 0e/0r (fallback)
- Archivos: BioD_RprGtA_310EntryAirlock (2/41 replies), BioD_CitHub_LOC_INT export 0

## 2. SpeakerList FName inner type off-by-8
- Impacto: speakers=0 en modo semántico (cithub_first_amb_a_dlg)
- Causa: resolveArrayCountAndPayloadStart lee instance number como count

## 3. payload_offset de arrays anidados incluye bytes del inner type
- Workaround: se deriva TargetEntryIDs de ReplyListNew inversa + raw byte scanning
