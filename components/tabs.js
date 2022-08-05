import { useState, useEffect } from "react";
import { useSelector, useDispatch, shallowEqual } from "react-redux";

import { equals_ignore_case } from "../utils";
import { ENVIRONMENT_ID } from "../reducers/types";

export default ({
  tabs,
  children,
}) => {
  const dispatch = useDispatch();
  const { environment } = useSelector(state => ({ environment: state.environment }), shallowEqual);
  const { environment_id } = { ...environment };

  const [openTab, setOpenTab] = useState(0);

  useEffect(() => {
    if (environment_id) {
      const index = tabs?.filter(t => !t?.hidden)
        .findIndex(t => equals_ignore_case(t?.title, environment_id));
      setOpenTab(index > -1 ? index : 0);
    }
  }, [environment_id]);

  const onClick = (tab, i) => {
    setOpenTab(i);
    dispatch({
      type: ENVIRONMENT_ID,
      value: tab?.title?.toLowerCase() || i,
    });
  };

  return (
    <div className="flex flex-wrap flex-col w-full tabs mt-4">
      <div className="flex lg:flex-wrap flex-row lg:space-x-2">
        {tabs?.filter(t => !t?.hidden).map((t, i) => (
          <div
            key={i}
            className="flex-none"
          >
            <button
              type="button"
              onClick={() => onClick(t, i)}
              className={openTab === i ? "tab tab-underline tab-active" : "tab tab-underline"}
            >
              {t.title}
            </button>
          </div>
        ))}
      </div>
      {tabs?.filter(t => !t?.hidden).map((t, i) => (
        <div
          key={i}
          className={`tab-content max-w-full ${openTab !== i ? "hidden" : "block"}`}
        >
          {t.content}
        </div>
      ))}
    </div>
  );
};