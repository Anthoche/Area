import React from "react";
import "./filtertag.css";

export default function FilterTag({ label, selected, onClick }) {
  return (
    <button
      className={`filter-tag ${selected ? "active" : ""}`}
      type="button"
      onClick={onClick}
    >
      {label}
    </button>
  );
}
