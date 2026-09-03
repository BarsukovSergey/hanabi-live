import jsdom from "jsdom";
import { CARD_H, CARD_W } from "../../../packages/client/src/game/ui/constants";
import Canvas2svg from "./lib/canvas2svg_node";

export function initCanvas(): {
  cvs: HTMLCanvasElement;
  ctx: CanvasRenderingContext2D;
} {
  const { document } = new jsdom.JSDOM("").window;
  const cvs = document.createElement("canvas");
  cvs.width = CARD_W;
  cvs.height = CARD_H;

  const ctx = new Canvas2svg({
    document,
    width: CARD_W,
    height: CARD_H,
  }) as unknown as CanvasRenderingContext2D;

  return {
    cvs,
    ctx,
  };
}

export function cloneCanvas(
  oldCvs: HTMLCanvasElement,
  oldCtx: CanvasRenderingContext2D,
): HTMLCanvasElement {
  // Keep the same adapter shape as the browser renderer. The Node implementation serializes the
  // drawing context to SVG instead of returning a real canvas element.
  void oldCvs;
  return (oldCtx as unknown as Canvas2svg).getSerializedSvg(
    true,
  ) as unknown as HTMLCanvasElement;
}

export function saveCanvas(
  cvs: HTMLCanvasElement,
  ctx: CanvasRenderingContext2D,
): HTMLCanvasElement {
  void cvs;
  return (ctx as unknown as Canvas2svg).getSerializedSvg(
    true,
  ) as unknown as HTMLCanvasElement;
}
