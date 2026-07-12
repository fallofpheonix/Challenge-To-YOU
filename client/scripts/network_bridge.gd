extends Node

signal fabric_updated(payload: Dictionary)
signal ontological_purge(reason: String)
signal archon_transmission(message: String)

var _socket: WebSocketPeer = WebSocketPeer.new()
var _is_connected: bool = false
const RIFT_URL = "ws://localhost:8080/rift"

func _ready() -> void:
	connect_to_rift()

func connect_to_rift(seed_val: int = -1, luck_val: float = 1.0, paradigm_val: String = "") -> void:
	if _socket.get_ready_state() != WebSocketPeer.STATE_CLOSED:
		_socket.close()
		_is_connected = false
		
	_socket = WebSocketPeer.new()
	
	var url = RIFT_URL
	if seed_val != -1 and paradigm_val != "":
		url += "?seed=%d&luck=%.2f&paradigm=%s" % [seed_val, luck_val, paradigm_val.to_upper()]
		
	_socket.connect_to_url(url)
	print("Connecting to: ", url)

func _process(_delta: float) -> void:
	_socket.poll()
	var state = _socket.get_ready_state()

	if state == WebSocketPeer.STATE_OPEN:
		if not _is_connected:
			_is_connected = true
			print("Rift connected.")
		while _socket.get_available_packet_count() > 0:
			_parse_inbound(_socket.get_packet().get_string_from_utf8())
	elif state == WebSocketPeer.STATE_CLOSED and _is_connected:
		_is_connected = false
		print("Rift disconnected.")

## Transmit a plain event with no payload (modify/exploit path).
func transmit_event(event_id: String) -> void:
	if not _is_connected:
		print("Cannot transmit — rift dormant.")
		return
	var msg = {"event": event_id}
	_socket.send_text(JSON.stringify(msg))

## Transmit an event with an attached payload (recognize, optimize, write_from_spec).
## payload is the player's answer string or full code block.
func transmit_event_with_payload(event_id: String, payload: String) -> void:
	if not _is_connected:
		print("Cannot transmit — rift dormant.")
		return
	var msg = {"event": event_id, "payload": payload}
	_socket.send_text(JSON.stringify(msg))

func _parse_inbound(raw_json: String) -> void:
	var json = JSON.new()
	if json.parse(raw_json) != OK:
		return
	var data: Dictionary = json.get_data()
	if data.get("type", "") == "archon_transmission":
		archon_transmission.emit(str(data.get("message", "")))
		return
		
	if data.has("error_message"):
		var err_msg: String = data["error_message"]
		if "PURGE" in err_msg:
			ontological_purge.emit(err_msg)
	fabric_updated.emit(data)
