export interface CardRendererBackend<Canvas, Image> {
  createCanvas: () => {
    cvs: Canvas;
    ctx: CanvasRenderingContext2D;
  };
  cloneCanvas: (canvas: Canvas, ctx: CanvasRenderingContext2D) => Image;
  saveCanvas: (canvas: Canvas, ctx: CanvasRenderingContext2D) => Image;
}
