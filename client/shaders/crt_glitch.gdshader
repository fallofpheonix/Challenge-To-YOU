shader_type canvas_item;

uniform sampler2D screen_texture : hint_screen_texture, repeat_disable, filter_nearest;

uniform float scanline_strength : hint_range(0.0, 1.0) = 0.15;
uniform float aberration : hint_range(0.0, 0.05) = 0.002;
uniform float glitch_intensity : hint_range(0.0, 1.0) = 0.0;
uniform float vignette_intensity : hint_range(0.0, 2.0) = 0.5;
uniform float curve_distortion : hint_range(0.0, 0.5) = 0.03;

// Simple random generator
float rand(vec2 co) {
    return fract(sin(dot(co.xy ,vec2(12.9898,78.233))) * 43758.5453);
}

// CRT lens distortion
vec2 curve(vec2 uv) {
    uv = (uv - 0.5) * 2.0;
    uv.x *= 1.0 + pow((abs(uv.y) * curve_distortion), 2.0);
    uv.y *= 1.0 + pow((abs(uv.x) * curve_distortion), 2.0);
    uv  = (uv / 2.0) + 0.5;
    return uv;
}

void fragment() {
    // 1. CRT Lens Curve distortion
    vec2 uv = curve(SCREEN_UV);
    
    // Check if UV is outside the screen bounds (vignette/border)
    if (uv.x < 0.0 || uv.x > 1.0 || uv.y < 0.0 || uv.y > 1.0) {
        COLOR = vec4(0.0, 0.0, 0.0, 1.0);
    } else {
        // 2. Glitch/Horizontal Tearing
        float time = TIME;
        if (glitch_intensity > 0.0) {
            // Generate pseudo-random offset for specific horizontal bands
            float glitch_y = floor(uv.y * 30.0 + time * 10.0);
            float glitch_trigger = rand(vec2(glitch_y, 123.45));
            if (glitch_trigger < glitch_intensity * 0.4) {
                float shift = (rand(vec2(glitch_y, 567.89)) - 0.5) * glitch_intensity * 0.08;
                uv.x += shift;
            }
        }
        
        // 3. Chromatic Aberration (Channel Split)
        float split = aberration;
        if (glitch_intensity > 0.0) {
            // Randomly spike aberration during glitches
            split += rand(vec2(time, 99.9)) * glitch_intensity * 0.015;
        }
        
        float r = texture(screen_texture, vec2(uv.x - split, uv.y)).r;
        float g = texture(screen_texture, uv).g;
        float b = texture(screen_texture, vec2(uv.x + split, uv.y)).b;
        
        vec4 color = vec4(r, g, b, 1.0);
        
        // 4. Scanlines
        float scanline = sin(uv.y * 400.0) * scanline_strength;
        color.rgb -= scanline;
        
        // 5. Vignette (Dark edges)
        float dist = distance(uv, vec2(0.5, 0.5));
        color.rgb *= (1.0 - dist * dist * vignette_intensity);
        
        COLOR = color;
    }
}
