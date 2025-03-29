import pytesseract


class LineData:
    def __init__(self, original_text:str, is_item:bool, name:str, price:int):
        self.original_text = original_text
        self.is_item = is_item
        self.name = name
        self.price = price
    
    def __str__(self):
        return f"{self.name}, {self.price} | {self.original_text}"


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

def extract_data_from_line(line:str) -> LineData|None:
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

if __name__ == "__main__":
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
