import { CHAIN_ID_DATA } from "./types";

export default function chain_id(
  state = {
    [`${CHAIN_ID_DATA}`]: null,
  },
  action
) {
  switch (action.type) {
    case CHAIN_ID_DATA:
      return {
        ...state,
        [`${CHAIN_ID_DATA}`]: action.value,
      };
    default:
      return state;
  };
};