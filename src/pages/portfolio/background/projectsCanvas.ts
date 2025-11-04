export const drawProjectsCanvasBackground = (canvas: HTMLCanvasElement | null) => {
  if (!canvas) return
  const ctx = canvas.getContext('2d')
  if (!ctx) return

  canvas.width = window.innerWidth
  canvas.height = window.innerHeight

  ctx.fillStyle = '#3498db'
  ctx.fillRect(0, 0, canvas.width, canvas.height)
}
