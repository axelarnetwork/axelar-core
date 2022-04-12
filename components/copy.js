import { useState, useEffect } from "react";
import { CopyToClipboard } from "react-copy-to-clipboard";
import { HiCheckCircle } from "react-icons/hi";
import { BiCopy } from "react-icons/bi";

export default ({ value, size = 16, onCopy, className = "" }) => {
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    const timeout = copied ? setTimeout(() => setCopied(false), 1 * 1000) : null;
    return () => clearTimeout(timeout);
  }, [copied, setCopied]);

  return copied ?
    <HiCheckCircle size={size} className={`text-green-300 dark:text-green-500 ${className}`} />
    :
    <CopyToClipboard
      text={value}
      onCopy={() => {
        setCopied(true);
        if (onCopy) {
          onCopy();
        }
      }}
    >
      <BiCopy size={size} className={`cursor-pointer text-gray-300 hover:text-gray-400 dark:text-gray-700 dark:hover:text-gray-600 ${className}`} />
    </CopyToClipboard>
};