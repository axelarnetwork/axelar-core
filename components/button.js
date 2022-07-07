export default ({
  buttonTitle,
  title,
  url,
  className,
  parentClassName,
}) => {
  buttonTitle = buttonTitle || title || "View";
  className = `bg-blue-500 hover:bg-blue-600 dark:bg-blue-600 dark:hover:bg-blue-500 rounded-lg no-underline text-white font-semibold py-2 px-2 ${className || ""}`;
  parentClassName = parentClassName || "pt-3";

  return (
    <div className={parentClassName}>
      {url ?
        <a
          href={url}
          target="_blank"
          rel="noopener noreferrer"
          title={title}
          className={className}
        >
          {buttonTitle}
        </a>
        :
        <button
          title={title}
          className={className}
        >
          {buttonTitle}
        </button>
      }
    </div>
  );
};