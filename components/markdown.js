import { useState, useEffect } from "react";
import Markdown from "markdown-to-jsx";

export default ({ src }) => {
  const [markdown, setMarkdown] = useState("");

  useEffect(() => {
    if (src) {
      fetch(src)
        .then(res => res.text())
        .then(md => setMarkdown(md));
    }
  }, [src]);

  return (
    <Markdown children={markdown} />
  );
};