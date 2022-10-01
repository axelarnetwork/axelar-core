import Copy from "../copy";
import resources from "../../data/resources.json";

const COLUMNS = [
  {
    id: "name",
    className: "align-top whitespace-nowrap font-semibold pt-3.5",
  },
  {
    id: "value",
  },
];

export default ({
  environment = "mainnet",
}) => {
  const _resources = resources?.[environment] ||
    [];

  return (
    <div className="mt-2">
      <table className="max-w-fit block border border-slate-200 dark:border-slate-800 rounded-lg overflow-x-auto">
        <tbody>
          {_resources
            .map((r, i) => {
              return (
                <tr
                  key={i}
                  className="border-none border-b"
                >
                  {COLUMNS
                    .map((c, j) => {
                      const {
                        id,
                        className,
                      } = { ...c };

                      const data = r?.[id];

                      return (
                        <th
                          key={j}
                          scope="col"
                          className={`${i % 2 === 0 ? "bg-transparent" : "bg-gray-50 dark:bg-black"} ${i === _resources.length - 1 ? j === 0 ? "rounded-bl-lg" : j === COLUMNS.length - 1 ? "rounded-br-lg" : "" : ""} border-none whitespace-nowrap py-3 px-4 ${className || ""}`}
                        >
                          {id === 'value' ?
                            <div className="flex flex-wrap items-center">
                              {(data || [])
                                .map((v, k) => {
                                  const {
                                    title,
                                    value,
                                  } = { ...v };

                                  const is_external = !value?.startsWith('/');

                                  return (
                                    <div
                                      key={k}
                                      className="flex items-center space-x-0.5 mb-2.5 mr-2.5"
                                    >
                                      <a
                                        href={value}
                                        target={is_external ?
                                          '_blank' :
                                          undefined
                                        }
                                        rel={is_external ?
                                          'noopener noreferrer' :
                                          undefined
                                        }
                                        className="bg-slate-100 dark:bg-slate-800 rounded-xl text-sm font-medium py-1 px-2.5"
                                      >
                                        {
                                          title ||
                                          value
                                        }
                                      </a>
                                      {
                                        !title &&
                                        (
                                          <Copy
                                            size={20}
                                            value={value}
                                          />
                                        )
                                      }
                                    </div>
                                  );
                                })
                              }
                            </div> :
                            data
                          }
                        </th>
                      );
                    })
                  }
                </tr>
              );
            })
          }
        </tbody>
      </table>
    </div>
  );
};