import SyntaxHighlighter from "react-syntax-highlighter";

export default ({
  language,
  children,
}) => {
  return (
    <SyntaxHighlighter
      language={language}
      className="code-block my-2"
    >
      {typeof children === "string" ?
        children.replace(/\\/g, "\\\n") :
        children
      }
    </SyntaxHighlighter>
  );
};