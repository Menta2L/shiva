package shader

var defaultTemplates = map[string]string{
	"cattributes": cattributes,
	"clights":     clights,
	"cmaterials":  cmaterials,
	"vbasic":      vbasic,
	"fbasic":      fbasic,
	"vstandard":   vstandard,
	"fstandard":   fstandard,
}

const cattributes = `{{ define "cattributes" }}
// Vertex attributes
layout(location = 0) in vec3  VertexPosition;
layout(location = 1) in vec3  VertexNormal;
layout(location = 2) in vec3  VertexColor;
layout(location = 3) in vec2  VertexTexcoord;
layout(location = 4) in float VertexDistance;
layout(location = 5) in vec4  VertexTexoffsets;
{{ end }}
`

const clights = `{{ define "clights" }}
{{ if .AmbientLightsMax }}
// Ambient lights uniforms
uniform vec3 AmbientLightColor[{{.AmbientLightsMax}}];
{{ end }}
{{ if .DirLightsMax }}
// Directional lights uniform array. Each directional light uses 2 elements
uniform vec3  DirLight[2*{{.DirLightsMax}}];
// Macros to access elements inside the DirectionalLight uniform array
#define DirLightColor(a)		DirLight[2*a]
#define DirLightPosition(a)		DirLight[2*a+1]
{{ end }}
{{ if .PointLightsMax }}
// Point lights uniform array. Each point light uses 3 elements
uniform vec3  PointLight[3*{{.PointLightsMax}}];
// Macros to access elements inside the PointLight uniform array
#define PointLightColor(a)			PointLight[3*a]
#define PointLightPosition(a)		PointLight[3*a+1]
#define PointLightLinearDecay(a)	PointLight[3*a+2].x
#define PointLightQuadraticDecay(a)	PointLight[3*a+2].y
{{ end }}
{{ if .SpotLightsMax }}
// Spot lights uniforms. Each spot light uses 5 elements
uniform vec3  SpotLight[5*{{.SpotLightsMax}}];
// Macros to access elements inside the PointLight uniform array
#define SpotLightColor(a)			SpotLight[5*a]
#define SpotLightPosition(a)		SpotLight[5*a+1]
#define SpotLightDirection(a)		SpotLight[5*a+2]
#define SpotLightAngularDecay(a)	SpotLight[5*a+3].x
#define SpotLightCutoffAngle(a)		SpotLight[5*a+3].y
#define SpotLightLinearDecay(a)		SpotLight[5*a+3].z
#define SpotLightQuadraticDecay(a)	SpotLight[5*a+4].x
{{ end }}
{{ end }}
`

const cmaterials = `{{ define "cmaterials" }}
// Material uniforms
uniform vec3	Material[6];
// Macros to access elements inside the Material uniform array
#define MatAmbientColor		Material[0]
#define MatDiffuseColor		Material[1]
#define MatSpecularColor	Material[2]
#define MatEmissiveColor	Material[3]
#define MatShininess		Material[4].x
#define MatOpacity			Material[4].y
#define MatPointSize		Material[4].z
#define MatPointRotationZ	Material[5].x
{{if .MaterialTexturesMax}}
// Textures uniforms
uniform sampler2D	MatTexture[{{.MaterialTexturesMax}}];
uniform mat3		MatTexinfo[{{.MaterialTexturesMax}}];
// Macros to access elements inside MatTexinfo uniform
#define MatTexOffset(a)		MatTexinfo[a][0].xy
#define MatTexRepeat(a)		MatTexinfo[a][1].xy
#define MatTexFlipY(a)		bool(MatTexinfo[a][2].x)
#define MatTexVisible(a)	bool(MatTexinfo[a][2].y)
{{ end }}
{{ end }}
`

const vbasic = `
{{ include "cattributes" }}
{{ include "cmaterials" }}
#version {{ .Version }}
{{ template "cattributes" . }}
{{ template "cmaterials" . }}
// Model uniforms
uniform mat4 MVP;
// Final output color for fragment shader
out vec3 Color;
void main() {
    Color = VertexColor;
    gl_Position = MVP * vec4(VertexPosition, 1.0);
}
`
const fbasic = `
#version {{.Version}}
in vec3 Color;
out vec4 FragColor;
void main() {
    FragColor = vec4(Color, 1.0);
}
`

const vstandard = `
#version {{.Version}}
{{template "cattributes" .}}
// Model uniforms
uniform mat4 ModelViewMatrix;
uniform mat3 NormalMatrix;
uniform mat4 MVP;
{{template "lights" .}}
{{template "cmaterial" .}}
{{template "phong_model" .}}
// Outputs for the fragment shader.
out vec3 ColorFrontAmbdiff;
out vec3 ColorFrontSpec;
out vec3 ColorBackAmbdiff;
out vec3 ColorBackSpec;
out vec2 FragTexcoord;
void main() {
    // Transform this vertex normal to camera coordinates.
    vec3 normal = normalize(NormalMatrix * VertexNormal);
    // Calculate this vertex position in camera coordinates
    vec4 position = ModelViewMatrix * vec4(VertexPosition, 1.0);
    // Calculate the direction vector from the vertex to the camera
    // The camera is at 0,0,0
    vec3 camDir = normalize(-position.xyz);
    // Calculates the vertex Ambient+Diffuse and Specular colors using the Phong model
    // for the front and back
    phongModel(position,  normal, camDir, MatAmbientColor, MatDiffuseColor, ColorFrontAmbdiff, ColorFrontSpec);
    phongModel(position, -normal, camDir, MatAmbientColor, MatDiffuseColor, ColorBackAmbdiff, ColorBackSpec);
    vec2 texcoord = VertexTexcoord;
    {{if .MatTexturesMax }}
    // Flips texture coordinate Y if requested.
    if (MatTexFlipY(0)) {
        texcoord.y = 1 - texcoord.y;
    }
    {{ end }}
    FragTexcoord = texcoord;
    gl_Position = MVP * vec4(VertexPosition, 1.0);
}
`

const fstandard = `
#version {{.Version}}
{{template "cmaterial" .}}
// Inputs from Vertex shader
in vec3 ColorFrontAmbdiff;
in vec3 ColorFrontSpec;
in vec3 ColorBackAmbdiff;
in vec3 ColorBackSpec;
in vec2 FragTexcoord;
// Output
out vec4 FragColor;
void main() {
    vec4 texCombined = vec4(1);
    // Combine all texture colors and opacity
    // Use Go templates to unroll the loop because non-const
    // array indexes are not allowed until GLSL 4.00.
    {{ range loop .MatTexturesMax }}
    if (MatTexVisible({{.}})) {
        vec4 texcolor = texture(MatTexture[{{.}}], FragTexcoord * MatTexRepeat({{.}}) + MatTexOffset({{.}}));
        if ({{.}} == 0) {
            texCombined = texcolor;
        } else {
            texCombined = mix(texCombined, texcolor, texcolor.a);
        }
    }
    {{ end }}
    vec4 colorAmbDiff;
    vec4 colorSpec;
    if (gl_FrontFacing) {
        colorAmbDiff = vec4(ColorFrontAmbdiff, MatOpacity);
        colorSpec = vec4(ColorFrontSpec, 0);
    } else {
        colorAmbDiff = vec4(ColorBackAmbdiff, MatOpacity);
        colorSpec = vec4(ColorBackSpec, 0);
    }
    FragColor = min(colorAmbDiff * texCombined + colorSpec, vec4(1));
}
`
