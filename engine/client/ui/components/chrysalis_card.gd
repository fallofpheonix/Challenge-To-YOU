extends PanelContainer

@export var header_text: String = ""
@export var body_text: String = ""
@export var accent_color: Color = Color(0, 1, 0.62, 1)

@onready var header_label: Label = $VBoxContainer/Header
@onready var body_label: RichTextLabel = $VBoxContainer/Body

func _ready() -> void:
	var theme_controller = get_node("/root/ChrysalisTheme")
	if theme_controller:
		theme_controller.apply_panel_style(self, theme_controller.colors.surface, accent_color * 0.5)

	header_label.text = header_text
	body_label.text = body_text

func set_header(text: String) -> void:
	header_text = text
	if header_label:
		header_label.text = text

func set_body(text: String) -> void:
	body_text = text
	if body_label:
		body_label.text = text

func set_accent(color: Color) -> void:
	accent_color = color
