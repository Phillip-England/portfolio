import { markdownToHtml } from "./marki";
import path from "node:path";

export class Article {
  path: string;
  number: string;
  name: string;
  html: string;
  dob: string;
  subText: string;
  constructor(
    path: string,
    number: string,
    name: string,
    html: string,
    subText: string,
    dob: string,
  ) {
    this.path = path;
    this.number = number;
    this.name = name;
    this.html = html;
    this.subText = subText;
    this.dob = dob;
  }
  static async new(pth: string) {
    let basename = path.basename(pth);
    let pathParts = basename.split(".");
    let number = pathParts[0];
    let name = snakecaseToPlainText(pathParts[1]);
    let html = await markdownToHtml(pth);
    let file = Bun.file(pth);
    let text = await file.text();
    let lines = text.split("\n");
    let metaData = {};
    let inMeta = true;
    let tripleLineCount = 0;
    let subText = "";
    let dob = "";
    for (let i = 0; i < lines.length; i++) {
      let line = lines[i].trim();
      if (line == "---") {
        tripleLineCount++;
        if (tripleLineCount == 2) {
          break;
        }
        continue;
      }
      if (line.startsWith("metaSub")) {
        let parts = line.split(":");
        let value = parts[1].trim();
        subText = value.slice(1, -1);
        continue;
      }
      if (line.startsWith("metaDob")) {
        let parts = line.split(":");
        let value = parts[1].trim();
        dob = value.slice(1, -1);
        continue;
      }
    }
    return new Article(pth, number, name, html, subText, dob);
  }
}
function snakecaseToPlainText(snakeCaseName: string): string {
  let parts = snakeCaseName.split("_");
  let plainTextParts: string[] = [];
  for (let i = 0; i < parts.length; i++) {
    let part = parts[i];
    if (part.length === 0) {
      continue;
    }
    let firstChar = part[0].toUpperCase();
    let restOfPart = part.substring(1);
    let capitalizedPart = firstChar + restOfPart;
    plainTextParts.push(capitalizedPart);
  }
  let plainText = plainTextParts.join(" ");
  return plainText;
}
