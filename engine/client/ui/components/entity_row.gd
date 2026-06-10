extends HBoxContainer

signal clicked

@export var entity_name: String = ""
@export var entity_type: String = ""
@export var status_text: String = ""
@export var status_color: Color = Color(0.5, 0.5, 0.5, 1)
@export var detail_text: String = ""

@onready var name_label: Label = $NameLabel
@onready var type_label: Label = $TypeLabel
@onready var status_indicator: ColorRect = $StatusIndicator
@onready var detail_label: Label = $DetailLabel

func _ready() -> void:
	name_label.text = entity_name
	type_label.text = entity_type
	status_indicator.color = status_color
	detail_label.text = detail_text

func _gui_input(event: InputEvent) -> void:
	if event is InputEventMouseButton and event.pressed and event.button_index == MOUSE_BUTTON_LEFT:
		clicked.emit()

func set_data(name_text: String, type_text: String, status: Color, detail: String) -> void:
	entity_name = name_text
	entity_type = type_text
	status_color = status
	detail_text = detail
	if name_label: name_label.text = name_text
	if type_label: type_label.text = type_text
	if status_indicator: status_indicator.color = status
	if detail_label: detail_label.text = detail
