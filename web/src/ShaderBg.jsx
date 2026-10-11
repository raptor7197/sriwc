import React, { useEffect, useRef } from 'react';

// WebGL fragment shader for Spider-Verse comic halftone background
const VERTEX_SHADER_SRC = `
  attribute vec2 position;
  void main() {
    gl_Position = vec4(position, 0.0, 1.0);
  }
`;

const FRAGMENT_SHADER_SRC = `
  precision highp float;
  uniform vec2 u_resolution;
  uniform float u_time;
  uniform vec2 u_mouse;

  // Ben-Day halftone comic dot raster
  float halftoneDot(vec2 uv, float angle, float scale, float radius) {
    float s = sin(angle);
    float c = cos(angle);
    mat2 rot = mat2(c, -s, s, c);
    vec2 p = rot * uv * scale;
    vec2 cell = fract(p) - 0.5;
    return 1.0 - smoothstep(radius - 0.08, radius + 0.08, length(cell));
  }

  // Spider-web radial lattice spokes & rings
  float spiderLattice(vec2 p, vec2 center, float time) {
    vec2 d = p - center;
    float r = length(d);
    float theta = atan(d.y, d.x);

    // 8 primary web spokes
    float spokes = abs(sin(theta * 8.0));
    float spokeLine = smoothstep(0.97, 0.995, spokes);

    // Concentric web tension rings drifting slowly
    float rings = abs(sin(r * 22.0 - time * 0.35));
    float ringLine = smoothstep(0.93, 0.985, rings);

    return (spokeLine * 0.5 + ringLine * 0.6) * smoothstep(1.6, 0.08, r);
  }

  void main() {
    vec2 uv = gl_FragCoord.xy / u_resolution.xy;
    vec2 p = (gl_FragCoord.xy - 0.5 * u_resolution.xy) / min(u_resolution.x, u_resolution.y);
    vec2 mouse = (u_mouse - 0.5 * u_resolution.xy) / min(u_resolution.x, u_resolution.y);

    float t = u_time * 0.4;

    // Palette: Deep Comic Ink, Spidey Crimson, Electric Cyan, Accent Yellow
    vec3 colInk = vec3(0.045, 0.05, 0.08);       // #0c0d14
    vec3 colCrimson = vec3(0.98, 0.12, 0.32);   // #fa1f52
    vec3 colElectric = vec3(0.0, 0.88, 1.0);    // #00e0ff
    vec3 colYellow = vec3(1.0, 0.85, 0.0);      // #ffd900

    // Subtle drift focal center
    vec2 focal = mix(vec2(sin(t * 0.25) * 0.2, cos(t * 0.2) * 0.2), mouse, 0.35);
    float distToFocal = length(p - focal);

    // Halftone screen print dots at classic comic screen angles (15 deg, 75 deg)
    float dotsCrimson = halftoneDot(uv + vec2(t * 0.008, 0.0), 0.26, 42.0, 0.26 + 0.12 * sin(p.x * 2.5 + t));
    float dotsCyan = halftoneDot(uv - vec2(0.0, t * 0.006), 1.30, 48.0, 0.22 + 0.10 * cos(p.y * 3.0 - t));

    // Spider web geometry
    float web = spiderLattice(p, focal, t);

    // Base dark ink
    vec3 col = colInk;

    // Halftone color accents
    col += colCrimson * dotsCrimson * 0.075;
    col += colElectric * dotsCyan * 0.065;

    // Web lattice glow
    col += colElectric * web * 0.14;
    col += colCrimson * web * 0.08 * smoothstep(0.8, 0.0, distToFocal);

    // Spider-sense chromatic pulse around cursor / focus
    float pulse = smoothstep(0.9, 0.0, distToFocal) * (0.05 + 0.03 * sin(t * 3.0));
    col += mix(colCrimson, colElectric, 0.5 + 0.5 * sin(t * 2.0)) * pulse;

    // Vignette
    float vig = smoothstep(1.5, 0.35, length(p));
    col *= vig;

    gl_FragColor = vec4(col, 1.0);
  }
`;

export default function ShaderBg() {
  const canvasRef = useRef(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;

    const gl = canvas.getContext('webgl', { antialias: false, powerPreference: 'low-power' });
    if (!gl) return;

    // Compile shader helper
    function createShader(type, source) {
      const shader = gl.createShader(type);
      gl.shaderSource(shader, source);
      gl.compileShader(shader);
      if (!gl.getShaderParameter(shader, gl.COMPILE_STATUS)) {
        gl.deleteShader(shader);
        return null;
      }
      return shader;
    }

    const vert = createShader(gl.VERTEX_SHADER, VERTEX_SHADER_SRC);
    const frag = createShader(gl.FRAGMENT_SHADER, FRAGMENT_SHADER_SRC);
    if (!vert || !frag) return;

    const program = gl.createProgram();
    gl.attachShader(program, vert);
    gl.attachShader(program, frag);
    gl.linkProgram(program);
    if (!gl.getProgramParameter(program, gl.LINK_STATUS)) {
      gl.deleteProgram(program);
      return;
    }

    gl.useProgram(program);

    // Quad covering viewport
    const positionBuffer = gl.createBuffer();
    gl.bindBuffer(gl.ARRAY_BUFFER, positionBuffer);
    gl.bufferData(
      gl.ARRAY_BUFFER,
      new Float32Array([
        -1, -1,
         1, -1,
        -1,  1,
        -1,  1,
         1, -1,
         1,  1,
      ]),
      gl.STATIC_DRAW
    );

    const posLoc = gl.getAttribLocation(program, 'position');
    gl.enableVertexAttribArray(posLoc);
    gl.vertexAttribPointer(posLoc, 2, gl.FLOAT, false, 0, 0);

    const uResolution = gl.getUniformLocation(program, 'u_resolution');
    const uTime = gl.getUniformLocation(program, 'u_time');
    const uMouse = gl.getUniformLocation(program, 'u_mouse');

    let animId;
    let mouse = [window.innerWidth * 0.5, window.innerHeight * 0.3];
    let startTime = performance.now();
    let isVisible = true;

    function handleResize() {
      if (!canvas) return;
      const dpr = Math.min(window.devicePixelRatio || 1, 1.5); // cap DPR for low GPU footprint
      const w = Math.floor(window.innerWidth * dpr);
      const h = Math.floor(window.innerHeight * dpr);
      if (canvas.width !== w || canvas.height !== h) {
        canvas.width = w;
        canvas.height = h;
        gl.viewport(0, 0, w, h);
      }
    }

    function handleMouseMove(e) {
      const dpr = Math.min(window.devicePixelRatio || 1, 1.5);
      mouse = [e.clientX * dpr, (window.innerHeight - e.clientY) * dpr];
    }

    function handleVisibility() {
      isVisible = !document.hidden;
      if (isVisible) {
        startTime = performance.now() - (lastElapsed * 1000);
        animId = requestAnimationFrame(render);
      } else {
        cancelAnimationFrame(animId);
      }
    }

    let lastElapsed = 0;

    function render(now) {
      if (!isVisible) return;
      const elapsed = (now - startTime) * 0.001;
      lastElapsed = elapsed;

      gl.uniform2f(uResolution, canvas.width, canvas.height);
      gl.uniform1f(uTime, elapsed);
      gl.uniform2f(uMouse, mouse[0], mouse[1]);

      gl.drawArrays(gl.TRIANGLES, 0, 6);
      animId = requestAnimationFrame(render);
    }

    handleResize();
    window.addEventListener('resize', handleResize);
    window.addEventListener('mousemove', handleMouseMove, { passive: true });
    document.addEventListener('visibilitychange', handleVisibility);

    animId = requestAnimationFrame(render);

    return () => {
      cancelAnimationFrame(animId);
      window.removeEventListener('resize', handleResize);
      window.removeEventListener('mousemove', handleMouseMove);
      document.removeEventListener('visibilitychange', handleVisibility);
      gl.deleteProgram(program);
      gl.deleteShader(vert);
      gl.deleteShader(frag);
      gl.deleteBuffer(positionBuffer);
    };
  }, []);

  return (
    <canvas
      ref={canvasRef}
      className="fixed inset-0 -z-10 h-full w-full pointer-events-none"
      style={{ background: '#0a0b12' }}
      aria-hidden="true"
    />
  );
}
