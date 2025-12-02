import React from "react";

export default function FilterTag({ label }) {
  return (
    <button className="filter-tag" type="button">
      {label}
    </button>
  );
}
