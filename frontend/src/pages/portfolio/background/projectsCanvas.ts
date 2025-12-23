export const drawProjectsCanvasBackground = (canvas: HTMLCanvasElement | null) => {
  if (!canvas) return

  const gl = canvas.getContext("webgl")
  if (!gl) return

  const resize = () => {
    canvas.width = window.innerWidth
    canvas.height = window.innerHeight
    gl.viewport(0, 0, canvas.width, canvas.height)
  }
  resize()
  window.addEventListener("resize", resize)

  const vertexSrc = `
    attribute vec2 position;
    void main() {
      gl_Position = vec4(position, 0.0, 1.0);
    }
  `

  const fragmentSrc = `
    precision mediump float;
    uniform vec2 u_resolution;
    uniform float u_time;

    float random(vec2 p){
      return fract(sin(dot(p, vec2(12.9898,78.233))) * 43758.5453);
    }

    float noise(vec2 p){
      vec2 i = floor(p);
      vec2 f = fract(p);
      float a = random(i);
      float b = random(i + vec2(1.0, 0.0));
      float c = random(i + vec2(0.0, 1.0));
      float d = random(i + vec2(1.0, 1.0));
      vec2 u = f * f * (3.0 - 2.0 * f);
      return mix(a, b, u.x) +
             (c - a)*u.y*(1.0-u.x) +
             (d - b)*u.x*u.y;
    }

    void main() {
      vec2 st = gl_FragCoord.xy / u_resolution.xy;
      st.x *= u_resolution.x / u_resolution.y;

      float t = u_time * 0.2;
      float n = noise(st * 4.0 + t);

      vec3 color1 = vec3(0.16, 0.27, 0.45); // 深い青
      vec3 color2 = vec3(0.4, 0.7, 1.0);    // 光の青
      vec3 color = mix(color1, color2, n);

      gl_FragColor = vec4(color, 1.0);
    }
  `

  const compile = (type: number, source: string) => {
    const shader = gl.createShader(type)!
    gl.shaderSource(shader, source)
    gl.compileShader(shader)
    if (!gl.getShaderParameter(shader, gl.COMPILE_STATUS)) {
      console.error(gl.getShaderInfoLog(shader))
    }
    return shader
  }

  const vertexShader = compile(gl.VERTEX_SHADER, vertexSrc)
  const fragmentShader = compile(gl.FRAGMENT_SHADER, fragmentSrc)

  const program = gl.createProgram()!
  gl.attachShader(program, vertexShader)
  gl.attachShader(program, fragmentShader)
  gl.linkProgram(program)
  gl.useProgram(program)

  const buffer = gl.createBuffer()
  gl.bindBuffer(gl.ARRAY_BUFFER, buffer)
  gl.bufferData(
    gl.ARRAY_BUFFER,
    new Float32Array([
      -1, -1,
       1, -1,
      -1,  1,
      -1,  1,
       1, -1,
       1,  1
    ]),
    gl.STATIC_DRAW
  )

  const position = gl.getAttribLocation(program, "position")
  gl.enableVertexAttribArray(position)
  gl.vertexAttribPointer(position, 2, gl.FLOAT, false, 0, 0)

  const resLoc = gl.getUniformLocation(program, "u_resolution")
  const timeLoc = gl.getUniformLocation(program, "u_time")

  let start = performance.now()
  const render = () => {
    const now = performance.now()
    const t = (now - start) / 1000

    gl.uniform2f(resLoc, canvas.width, canvas.height)
    gl.uniform1f(timeLoc, t)

    gl.drawArrays(gl.TRIANGLES, 0, 6)
    requestAnimationFrame(render)
  }

  render()
}
