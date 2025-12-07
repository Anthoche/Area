import React from "react";
import "./filtertag.css";

export default function FilterTag({ label }) {
  return (
    <button className="filter-tag" type="button">
      {label}
    </button>
  );
}
