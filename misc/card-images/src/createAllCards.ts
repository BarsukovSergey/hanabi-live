// Generate the canonical card images by running the same renderer that the web client uses.

import { getVariant } from "@hanabi-live/game";
import fs from "node:fs";
import path from "node:path";
import { drawCards } from "../../../packages/client/src/game/ui/drawCards";
import * as drawCardsNode from "./drawCardsNode";

const variantName = process.argv[2] ?? "No Variant";
const outputDirectory = path.resolve(
  process.argv[3] ?? path.join(__dirname, "cards"),
);

const variant = getVariant(variantName);
const cardImages = drawCards(
  variant,
  false,
  false,
  false,
  drawCardsNode.initCanvas,
  drawCardsNode.cloneCanvas,
  drawCardsNode.saveCanvas,
);

fs.rmSync(outputDirectory, { recursive: true, force: true });
fs.mkdirSync(outputDirectory, { recursive: true });

for (const [key, value] of cardImages) {
  const filePath = path.join(outputDirectory, `${key}.svg`);
  fs.writeFileSync(filePath, value as unknown as string, "utf8");
}

console.log(
  `Created ${cardImages.size} card images for \"${variant.name}\" in: ${outputDirectory}`,
);
