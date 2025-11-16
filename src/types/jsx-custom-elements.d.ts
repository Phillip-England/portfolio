export {};

declare global {
  namespace JSX {
    interface IntrinsicElements {
      "title-links": {
        target?: string;
        "link-class"?: string;
        "link-wrapper-class"?: string;
        offset?: string | number;
        [key: string]: any;
      };
    }
  }
}
