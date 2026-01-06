import React from "react";
import "./searchbar.css";

export default function SearchBar({ value, onChange, placeholder = "Search" }) {
    return (
      <div className="searchbar-wrapper">
        <input
          className="search-input"
          placeholder={placeholder}
          value={value}
          onChange={(e) => onChange?.(e.target.value)}
        />
      </div>
    );
}
