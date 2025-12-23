export const drawIntroCanvasBackground = (canvas: HTMLCanvasElement | null) => {
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

  const compile = (type: number, source: string) => {
    const shader = gl.createShader(type)!
    gl.shaderSource(shader, source)
    gl.compileShader(shader)
    if (!gl.getShaderParameter(shader, gl.COMPILE_STATUS))
      console.error(gl.getShaderInfoLog(shader))
    return shader
  }

  const vertexBG = `
    attribute vec2 position;
    void main(){
      gl_Position = vec4(position, 0.0, 1.0);
    }
  `

  const fragmentBG = `
    precision mediump float;
    uniform vec2 u_resolution;
    uniform float u_time;

    vec3 hsv2rgb(float h,float s,float v){
      vec3 rgb = clamp(abs(mod(h*6.0+vec3(0,4,2),6.0)-3.0)-1.0,0.0,1.0);
      rgb = rgb*rgb*(3.0-2.0*rgb);
      return v*mix(vec3(1.0),rgb,s);
    }

    void main(){
      vec2 st = gl_FragCoord.xy / u_resolution.xy;
      st.x *= u_resolution.x/u_resolution.y;
      float t = u_time*0.6;

      float wave = sin(st.x*10.0+t)+sin(st.y*10.0-t);
      float swirl = sin((st.x+st.y)*6.0+t*1.5);
      float glow = pow(abs(wave+swirl),2.0);

      float hue = fract((st.x+st.y)*0.6+t*0.2);
      vec3 col = hsv2rgb(hue,1.0,1.0);
      col *= 0.4 + glow * 1.5;
      gl_FragColor = vec4(col,1.0);
    }
  `

  const bgProg = gl.createProgram()!
  gl.attachShader(bgProg, compile(gl.VERTEX_SHADER, vertexBG))
  gl.attachShader(bgProg, compile(gl.FRAGMENT_SHADER, fragmentBG))
  gl.linkProgram(bgProg)

  const quad = gl.createBuffer()
  gl.bindBuffer(gl.ARRAY_BUFFER, quad)
  gl.bufferData(gl.ARRAY_BUFFER, new Float32Array([
    -1,-1, 1,-1, -1,1,
    -1,1, 1,-1, 1,1
  ]), gl.STATIC_DRAW)

  const bgPos = gl.getAttribLocation(bgProg, "position")
  const uRes = gl.getUniformLocation(bgProg, "u_resolution")
  const uTime = gl.getUniformLocation(bgProg, "u_time")

  const vertexPoly = `
    attribute vec2 pos;
    void main(){
      gl_Position = vec4(pos,0.0,1.0);
    }
  `

  const fragmentPoly = `
    precision mediump float;

    uniform vec2 u_resolution;
    uniform vec2 u_sites[60];
    uniform int u_count;

    void main(){
      vec2 st = gl_FragCoord.xy / u_resolution.xy;
      st.x *= u_resolution.x / u_resolution.y;
      
      float d1 = 9999.0;
      float d2 = 9999.0;
      int index = 0;

      for(int i = 0; i < 60; i++){
        vec2 diff = st - u_sites[i];
        float d = dot(diff, diff);

        if(d < d1) {
          d2 = d1;
          d1 = d;
          index = i;
        } else if(d < d2) {
          d2 = d;
        }
      }

      float r1 = sqrt(d1);
      float r2 = sqrt(d2);
      
      float edgePx = (r2 - r1) * min(u_resolution.x, u_resolution.y);
      float edgeWidthPx = 2.0;
      float feather     = 1.5;
      float line = smoothstep(edgeWidthPx, edgeWidthPx + feather, edgePx);

      float t = float(index) / float(u_count);
      vec3 col(0.0, 0.0, 0.0);

      float fillAlpha = 1.0;
      float edgeAlpha = 0.0;

      float alpha = mix(edgeAlpha, fillAlpha, line);
      gl_FragColor = vec4(col, alpha);
    }
  `

  const polyProg = gl.createProgram()!
  gl.attachShader(polyProg, compile(gl.VERTEX_SHADER, vertexPoly))
  gl.attachShader(polyProg, compile(gl.FRAGMENT_SHADER, fragmentPoly))
  gl.linkProgram(polyProg)

  const POLY_COUNT = 60
  const sites = new Float32Array(POLY_COUNT * 2)

  for(let i = 0; i < POLY_COUNT; i++){
    sites[i*2]   = Math.random()*2
    sites[i*2+1] = Math.random()*2
  }

  const polyBuf = gl.createBuffer()
  gl.bindBuffer(gl.ARRAY_BUFFER, polyBuf)
  gl.bufferData(gl.ARRAY_BUFFER, new Float32Array([
    -1,-1, 1,-1, -1,1,
    -1,1, 1,-1, 1,1
  ]), gl.STATIC_DRAW)

  const polyPos = gl.getAttribLocation(polyProg,"pos")
  const uResPoly = gl.getUniformLocation(polyProg,"u_resolution")
  const uSites   = gl.getUniformLocation(polyProg,"u_sites")
  const uCount   = gl.getUniformLocation(polyProg,"u_count")

  gl.enable(gl.BLEND)
  gl.blendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

  const start = performance.now()

  const render = () => {
    const t = (performance.now() - start) / 1000

    // --- background
    gl.useProgram(bgProg)
    gl.bindBuffer(gl.ARRAY_BUFFER, quad)
    gl.enableVertexAttribArray(bgPos)
    gl.vertexAttribPointer(bgPos,2,gl.FLOAT,false,0,0)
    gl.uniform2f(uRes, canvas.width, canvas.height)
    gl.uniform1f(uTime, t)
    gl.drawArrays(gl.TRIANGLES,0,6)

    // --- Voronoi
    gl.useProgram(polyProg)
    gl.bindBuffer(gl.ARRAY_BUFFER, polyBuf)
    gl.enableVertexAttribArray(polyPos)
    gl.vertexAttribPointer(polyPos,2,gl.FLOAT,false,0,0)
    gl.uniform2f(uResPoly, canvas.width, canvas.height)
    gl.uniform1i(uCount, POLY_COUNT)
    gl.uniform2fv(uSites, sites)
    gl.drawArrays(gl.TRIANGLES,0,6)

    requestAnimationFrame(render)
  }

  render()
}
