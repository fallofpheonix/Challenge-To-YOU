extends Node

var socket := WebSocketPeer.new()
var ws_url := "ws://127.0.0.1:8080/telemetry"

signal packet_received(type: String, data: Dictionary)

func _ready():
	set_process(false)
	connect_to_core()

func connect_to_core():
	var err = socket.connect_to_url(ws_url)
	if err != nil:
		print("[Network Error] Failed to map socket: ", err)
		return
	set_process(true)
	print("[Network] Connecting to Chrysalis Core...")

func _process(_delta):
	socket.poll()
	var state = socket.get_ready_state()
	
	if state == WebSocketPeer.STATE_OPEN:
		while socket.get_available_packet_count() > 0:
			var raw_packet := socket.get_packet().get_string_from_utf8()
			_parse_and_route_packet(raw_packet)
	elif state == WebSocketPeer.STATE_CLOSED:
		set_process(false)
		print("[Network] Core socket disconnected. Re-connecting...")
		await get_tree().create_timer(2.0).timeout
		connect_to_core()

func send_command(type: String, command_payload: Dictionary):
	if socket.get_ready_state() == WebSocketPeer.STATE_OPEN:
		var out_packet = {
			"packet_type": type,
			"tick": 0, # Evaluated at arrival point
			"payload": command_payload
		}
		socket.send_text(JSON.stringify(out_packet))

func _parse_and_route_packet(raw: String):
	var json = JSON.new()
	if json.parse(raw) == OK:
		var data = json.get_data()
		if data is Dictionary:
			packet_received.emit(data.get("packet_type", ""), data)
