extends Node

const ChrysalisColors = preload("res://ui/theme/chrysalis_colors.gd")
var colors = ChrysalisColors.DARK_PALETTE

func _ready() -> void:
	pass

func color(name: String) -> Color:
	return colors.get(name, Color.WHITE)

func apply_panel_style(panel: Control, bg_color: Color = colors.surface, border_color: Color = colors.border) -> void:
	var style = StyleBoxFlat.new()
	style.bg_color = bg_color
	style.border_color = border_color
	style.border_width_bottom = 1
	style.border_width_top = 1
	style.border_width_left = 1
	style.border_width_right = 1
	style.corner_radius_bottom_left = 4
	style.corner_radius_bottom_right = 4
	style.corner_radius_top_left = 4
	style.corner_radius_top_right = 4
	style.content_margin_left = 8
	style.content_margin_right = 8
	style.content_margin_top = 8
	style.content_margin_bottom = 8
	panel.add_theme_stylebox_override("panel", style)

func apply_button_style(btn: Button, bg_color: Color = colors.surface_alt, text_color: Color = colors.accent) -> void:
	var normal = StyleBoxFlat.new()
	normal.bg_color = bg_color
	normal.border_color = colors.border
	normal.border_width_bottom = 1
	normal.border_width_top = 1
	normal.border_width_left = 1
	normal.border_width_right = 1
	normal.corner_radius_bottom_left = 4
	normal.corner_radius_bottom_right = 4
	normal.corner_radius_top_left = 4
	normal.corner_radius_top_right = 4
	btn.add_theme_stylebox_override("normal", normal)
	btn.add_theme_color_override("font_color", text_color)
	btn.add_theme_color_override("font_hover_color", colors.primary_text)
	btn.add_theme_color_override("font_pressed_color", colors.accent_dim)

	var hover = normal.duplicate()
	hover.border_color = colors.border_focus
	hover.bg_color = colors.surface
	btn.add_theme_stylebox_override("hover", hover)

	var pressed = normal.duplicate()
	pressed.bg_color = colors.secondary_background
	pressed.border_color = colors.accent
	btn.add_theme_stylebox_override("pressed", pressed)

func apply_label_style(label: Label, text_color: Color = colors.primary_text, font_size: int = 14) -> void:
	label.add_theme_color_override("font_color", text_color)
	label.add_theme_font_size_override("font_size", font_size)
