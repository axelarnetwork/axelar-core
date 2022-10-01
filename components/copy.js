import { useState, useEffect } from "react"
import { CopyToClipboard } from "react-copy-to-clipboard"
import { HiCheckCircle } from "react-icons/hi"
import { IoMdCopy } from "react-icons/io"

export default ({
  size = 18,
  value,
  title,
  onCopy,
  className = "",
}) => {
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    const timeout = copied ?
      setTimeout(() =>
        setCopied(false),
        1 * 1000,
      ) :
      undefined;

    return () => clearTimeout(timeout);
  }, [copied])

  return copied ?
    <div className={`${title ? "min-w-max" : ""} flex items-center space-x-1`}>
      {title && (
        <span>
          {title}
        </span>
      )}
      <HiCheckCircle
        size={size}
        className={
          className ||
          "text-green-400 dark:text-green-300"
        }
      />
    </div> :
    <CopyToClipboard
      text={value}
      onCopy={() => {
        setCopied(true)

        if (onCopy) {
          onCopy()
        }
      }}
    >
      <div className={`${title ? "min-w-max" : ""} flex items-center space-x-1`}>
        {title && (
          <span>
            {title}
          </span>
        )}
        <IoMdCopy
          size={size}
          className={
            className ||
            "cursor-pointer text-slate-300 hover:text-slate-400 dark:text-slate-700 dark:hover:text-slate-600"
          }
        />
      </div>
    </CopyToClipboard>
};