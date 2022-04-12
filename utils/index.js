export default {
  ellipse: (string, length = 10) => !string ? "" : string.length < (length * 2) + 3 ? string : `${string.slice(0, length)}...${string.slice(-length)}`,
};