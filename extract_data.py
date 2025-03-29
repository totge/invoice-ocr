import pytesseract
import cv2
import numpy as np
import json

class InvoiceData:
    __slots__ = ["raw_lines", "items", "total"]

    def __init__(self, path:str):
        self.items:list[LineData] = []
        self.raw_lines:list[str] = []

        self._read_invoice(path)
        item_lines = self._get_item_lines()
        self._process_items(item_lines)

        self._extract_total()

    def _read_invoice(self, path:str):
        """
        Extracts text from the invoice in the provided path

        @param path: path to the invoice
        """

        text_raw:str = pytesseract.image_to_string(path)
    
        lines = [line.strip() for line in text_raw.split("\n")]

        self.raw_lines = lines


    def _get_item_lines(self) -> list[str]:
        # filtering empty lines
        clean_lines = [line for line in self.raw_lines if len(line) != 0]

        item_start_idx:int = 0
        item_end_idx:int = -1

        # finding the first and last item line
        for i, line in enumerate(clean_lines):
            if "ertekesitesi bizonylat" in line.lower():
                item_start_idx = i
            
            if "fizetend" in line.lower():
                item_end_idx = i

        return clean_lines[item_start_idx:item_end_idx]

    def _process_items(self, item_lines:list[str]):
        for line in item_lines:
            parts = line.split(" ")

            # filter out lines with unit price
            if len(parts[-1]) != 3 or "/" in parts[-1]:
                continue
                # return LineData(line, False, None, None)
            
            name_pieces = []
            price_pieces = []

            for part in parts:
                try:
                    int(part)
                    price_pieces.append(part)
                except Exception:
                    name_pieces.append(part)


            name = " ".join(name_pieces[:-1])

            # TODO: handle better this error
            try:
                price = int("".join(price_pieces))
            except:
                price = -1
            
            item = LineData(original_text=line, 
                            is_item=True, 
                            name=name, 
                            fullprice=price, 
                            discount=0)
            
            self.items.append(item) 

    def _extract_total(self):
        total_text_parts:list[str] = []

        for line in self.raw_lines:
            parts = line.split(" ")

            if len(parts) < 2:
                continue
                
            if "fizetend" in parts[0]:
                for part in parts[1:]:
                    try:
                        int(part)
                        total_text_parts.append(part)
                    except Exception:
                        continue
        
        total_text = "".join(total_text_parts)
        
        # TODO: fix this to handle errors better
        try:
             
            self.total = int(total_text)
        except Exception:
            self.total = -1

    def calculate_total(self) -> int:
        total = 0
        for item in self.items:
            total += item.price

        return total

    def to_json(self) -> str:
        """
        Returns the object's data in a json string.
        """
        items = [{"name": item.name, "price": item.price, "parsed_line": item.original_text} for item in self.items]

        data = {
            # TODO: parse dat from the invoice
            "datetime": "date",
            "items": items,
            "parsed_total": self.total,
            "calculated_total": self.calculate_total()
        }

        return json.dumps(data, indent=4)

class LineData:
    def __init__(self, original_text:str, is_item:bool, name:str, fullprice:int, discount: int):
        self.original_text = original_text
        self.is_item = is_item
        self.name = name
        self.fullprice = fullprice
        self.discount = discount
        self.price = fullprice + discount
    
    def __str__(self):
        return f"{self.name}, {self.price} | {self.original_text}"
    
    def add_discount(self, discount:int):
        """
        Adds a discount to the item, sets the the discount property to the new value and recalculates the price.

        @param discount: discount amount (as a negative int)
        """
        self.discount = discount
        self.price = self.fullprice + self.discount


def get_lines(path: str) -> list[str]:
    text_raw:str = pytesseract.image_to_string(path)
    
    lines = [line.strip() for line in text_raw.split("\n")]

    return lines

def filter_lines(lines:list[str]) -> list[str]:
    # filtering empty lines
    clean_lines = [line for line in lines if len(line) != 0]

    item_start_idx:int = 0
    item_end_idx:int = -1

    for i, line in enumerate(clean_lines):
        if "ertekesitesi bizonylat" in line.lower():
            item_start_idx = i
        
        if "fizetend" in line.lower():
            item_end_idx = i

    return clean_lines[item_start_idx:item_end_idx]

def extract_data_from_line(self, line:str) -> LineData|None:
    parts = line.split(" ")

    # filter out lines with unit price
    if len(parts[-1]) != 3 or "/" in parts[-1]:
        return LineData(line, False, None, None)
    
    name_pieces = []
    price_pieces = []

    for part in parts:
        try:
            int(part)
            price_pieces.append(part)
        except Exception:
            name_pieces.append(part)


    name = " ".join(name_pieces[:-1])
    price = int("".join(price_pieces))
    
    return LineData(line, True, name, price)

def main():
    lines = get_lines("test_invoices/16000333862025032623918.png")

    lines = filter_lines(lines)
    #print(lines)
    total_price = 0
    for line in lines:
        data = extract_data_from_line(line)

        print(data)

        if data.is_item:
            total_price += data.price
            # print(data)
    print(total_price)


def image_cut():
    # Example usage
    image_path = "test_invoices/16000333892025031436070 (1).png"  # Replace with your image path
    sections = split_invoice_by_text_lines(image_path)

    # Save or display the sections
    for idx, sec in enumerate(sections):
        cv2.imwrite(f"test_invoices/split_image/section_{idx}.png", sec)
    
if __name__ == "__main__":

    # image_cut()
    # main()

    invoice_paths = [
        "test_invoices/16000333862025032623918.png",  
        "test_invoices/16000335892025032535080.png",
        "test_invoices/16000333892025031436070 (1).png",
        "test_invoices/16000335892025032735441.png"
    ]

    for inv_path in invoice_paths:

        invoice = InvoiceData(inv_path)

        print(invoice.to_json())
    

