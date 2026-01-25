"use client";

import { Button } from "@/components/shadcn/button";
import {
  Command,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "@/components/shadcn/command";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/shadcn/popover";
import { cn } from "@/lib/utils";
import { CheckIcon, ChevronsUpDownIcon } from "lucide-react";
import { useState } from "react";

interface SearchableSelectProps<T> {
  items: T[];
  value: T | null;
  onSelect: (item: T) => void;
  getKey: (item: T) => string;
  getLabel: (item: T) => string;
  getSearchValue?: (item: T) => string;
  placeholder?: string;
  searchPlaceholder?: string;
  emptyMessage?: string;
  disabled?: boolean;
  renderItem?: (item: T, isSelected: boolean) => React.ReactNode;
  isItemDisabled?: (item: T) => boolean;
  className?: string;
}

export function SearchableSelect<T>({
  items,
  value,
  onSelect,
  getKey,
  getLabel,
  getSearchValue,
  placeholder = "Select...",
  searchPlaceholder = "Search...",
  emptyMessage = "No results found.",
  disabled = false,
  renderItem,
  isItemDisabled,
  className,
}: SearchableSelectProps<T>) {
  const [open, setOpen] = useState(false);

  const handleSelect = (item: T) => {
    if (isItemDisabled?.(item)) return;
    onSelect(item);
    setOpen(false);
  };

  return (
    <Popover open={open} onOpenChange={setOpen}>
      <PopoverTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          disabled={disabled}
          className={cn("w-full justify-between font-normal", className)}
        >
          {value ? getLabel(value) : placeholder}
          <ChevronsUpDownIcon className="ml-2 size-4 shrink-0 opacity-50" />
        </Button>
      </PopoverTrigger>
      <PopoverContent className="w-[var(--radix-popover-trigger-width)] p-0" align="start">
        <Command>
          <CommandInput placeholder={searchPlaceholder} />
          <CommandList>
            <CommandEmpty>{emptyMessage}</CommandEmpty>
            <CommandGroup>
              {items.map((item) => {
                const key = getKey(item);
                const isSelected = value ? getKey(value) === key : false;
                const itemDisabled = isItemDisabled?.(item) ?? false;

                return (
                  <CommandItem
                    key={key}
                    value={getSearchValue?.(item) ?? getLabel(item)}
                    onSelect={() => handleSelect(item)}
                    disabled={itemDisabled}
                    className={cn(itemDisabled && "opacity-50")}
                  >
                    {renderItem ? (
                      renderItem(item, isSelected)
                    ) : (
                      <>
                        <CheckIcon
                          className={cn("mr-2 size-4", isSelected ? "opacity-100" : "opacity-0")}
                        />
                        {getLabel(item)}
                      </>
                    )}
                  </CommandItem>
                );
              })}
            </CommandGroup>
          </CommandList>
        </Command>
      </PopoverContent>
    </Popover>
  );
}
