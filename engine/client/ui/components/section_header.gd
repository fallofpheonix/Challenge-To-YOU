extends HBoxContainer

@export var section_title: String = ""

@onready var title_label: Label = $TitleLabel
@onready var accent_bar: ColorRect = $AccentBar

func _ready() -> void:
	title_label.text = section_title
	title_label.add_theme_color_override("font_color", Color(0, 1, 0.62, 1))
	title_label.add_theme_font_size_override("font_size", 16)

func set_title(text: String) -> void:
	section_title = text
	if title_label:
		title_label.text = text
