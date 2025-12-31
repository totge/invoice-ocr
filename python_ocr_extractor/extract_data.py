import pytesseract
import cv2
import json
import re
import bisect
from datetime import datetime

# max 'left' value for a line to be considered not indented
LINE_START_LIMIT = 100
# treshold for max difference in 'top' value of words in the same line
LINE_TOP_TRESHOLD = 15
# min 'left' value for a word to be considered as a part of the price
PRICE_MIN_LEFT_BOUNDARY = 720
# min 'left' value for a word to be considered as a part of the date
DATE_MIN_LEFT_BOUNDARY = 620




def main():
    invoice_paths = [
            "test_invoices/16000333862025032623918.png",  
            "test_invoices/16000335892025032535080.png",
            "test_invoices/16000333892025031436070 (1).png",
            "test_invoices/16000335892025032735441.png"
        ]

    for inv_path in invoice_paths:
        data_dict = process_invoice(inv_path)
        # print(data_dict)
        invoice = Invoice(data_dict)

        print(invoice.to_json())
        # invoice._clean_lines()

def process_invoice(path:str) -> dict:
    """
    Extracts tesseract data from the image at the provided path, returns a dict with the data

    @param path: path to the image
    """
    img = preprocess_image(path)
    custom_config = custom_config = r'--psm 6'
    data_dict = pytesseract.image_to_data(img, lang="eng", output_type=pytesseract.Output.DICT ,config=custom_config)

    return data_dict

def preprocess_image(path:str) -> cv2.typing.MatLike:
    """
    Runs the preprocessing steps on the image at the provided path

    @param path: path to the image
    """
    image = cv2.imread(path)
    # graycsaling the image
    gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
    # thresholding -> making it black and white
    _, bw = cv2.threshold(gray, 180, 255, cv2.THRESH_BINARY)
    return bw


class ExtractedWord:
    def __init__(self, text:str, left:int, top:int, width:int, height:int):
        self.text = text
        self.left = left
        self.top = top
        self.width = width
        self.height = height

    def __str__(self):
        return f"{self.text} at top: {self.top} left:{self.left}"
    
    def __repr__(self):
        return f"{self.text} at top: {self.top} left:{self.left}"
    
    @property
    def bottom(self)->int:
        return self.top + self.height
    
    @property
    def right(self)->int:
        return self.left + self.width

class Line:
    def __init__(self, line_num:int):
        self.line_num = line_num
        self.words:list[ExtractedWord] = []

    def __str__(self):
        texts = [word.text for word in self.words]
        return " ".join(texts)
    
    @property
    def line_start(self):
        return min([word.left for word in self.words])
    
    @property
    def line_top(self):
        return min([word.top for word in self.words])
    
    @property
    def line_bottom(self):
        return max([word.bottom for word in self.words])
    
    @property
    def text(self):
        word_text = [word.text for word in self.words]
        return " ".join(word_text)
    
    @property
    def avg_line_top(self) -> float:
        if len(self.words) == 0:
            return 0

        sum_top = sum([word.top for word in self.words])
        if len(self.words) < 3:
            avg = sum_top / len(self.words)
        else:
            min_top = min([word.top for word in self.words])
            max_top = max([word.top for word in self.words])
            sum_corr = sum_top - min_top - max_top
            avg = sum_corr / (len(self.words) - 2)
        
        return avg

    def add_word(self, word:ExtractedWord):
        if len(self.words) == 0:
            self.words.append(word)
            return

        bisect.insort_right(self.words, word, key=lambda x: x.left)

    def is_position_occupied(self, position:int):
        for word in self.words:
            if word.left <= position <= word.right:
                return True
            
        return False
    
    def extract_item_data(self) -> tuple[str|None, int|None]:
        """
        It tries to parse the line and extract the order item name and price from it, if unsuccessful, it returns a None for both.
        """

        name_parts:list[str] = []
        price_parts:list[str] = []

        # we expect at least 3 words in every order line: name, price, VAT type
        if len(self.words) < 3:
            return None, None

        # we don't include the trailing VAT type indicator
        for word in self.words[:-1]:
            if word.left < PRICE_MIN_LEFT_BOUNDARY:
                name_parts.append(word.text)
            else:
                # only extract numeric characters from price text parts
                price_parts.extend(re.findall(r'\d+', word.text))

        
        item_name = " ".join(name_parts)
        try:
            item_price = int("".join(price_parts))
        except:
            return None,None

        return item_name, item_price

class Item:
    __slots__ = ["line_num", "name", "full_price", "discount", "raw_text"]

    # TODO: add documentation
    def __init__(self, line_num:int, name:str, full_price:int, discount:int, raw_text:str):
        """
            @param line_num:
        """
        self.line_num = line_num
        self.name = name
        self.full_price = full_price
        self.discount = discount
        self.raw_text = raw_text

    @property
    def price(self):
        return self.full_price + self.discount
    
    def add_discount(self, discount:int):
        self.discount = discount

        

class Invoice:
    __slots__ = ["raw_data", "_words", "_lines", "order_items", "invoice_total", "invoice_date"]
    def __init__(self, raw_data:dict):
        """
        Provides a high level object to access information from an invoice

        @param raw_data: dictionary containing raw data coming from tesseract
        """
        self.raw_data = raw_data
        self._words:list[ExtractedWord] = []
        self._lines:list[Line] = []
        self.order_items:list[Item] = []
 
        # data processing steps to build up the object
        self._extract_words()
        self._build_lines()
        self._clean_lines()
        self._extract_order_items()

        # extracting additional data from the invoice
        self.invoice_total = self._extract_total()
        self.invoice_date = self._extract_date()
    
    def _extract_words(self):
        for i in range(len(self.raw_data["text"])):
            if self.raw_data["text"][i] == "":
                continue
            
            word = ExtractedWord(
                self.raw_data["text"][i],
                self.raw_data["left"][i],
                self.raw_data["top"][i],
                self.raw_data["width"][i],
                self.raw_data["height"][i]
            )

            self._words.append(word)
        
        self._words.sort(key=lambda x: x.top)
    
    def _build_lines(self):
        line_num = 0
        line = Line(line_num)

        line.add_word(self._words[0])

        for i in range(1, len(self._words), 1):
            if abs(self._words[i].top - line.avg_line_top) > LINE_TOP_TRESHOLD:
                self._lines.append(line)
                line_num += 1
                line = Line(line_num)
            
            line.add_word(self._words[i])
        self._lines.append(line)

    def _get_order_items_area(self):
        # determining area to look for order items
        # TODO: make this more robust
        upper_boundary = 0
        lower_boundary = 0

        for line in self._lines:
            if "ertekesitesi bizonylat" in line.text.lower():
                upper_boundary = line.line_bottom

            if "fizetend" in line.text.lower() or "osszesen" in line.text.lower():
                lower_boundary = line.line_top

        return upper_boundary, lower_boundary
    
    # TODO: maybe also need a rule, if the word to delete is longer than 1-2 then keep it
    def _clean_lines(self):
        top_limit, bottom_limit = self._get_order_items_area()

        for line in self._lines:
            # check if line is in orders area
            if line.line_top < top_limit or line.line_bottom > bottom_limit:
                continue
            
            # check is line is not indented
            if line.line_start > LINE_START_LIMIT:
                continue
            
            sus_word_idx = []
            for i, word in enumerate(line.words):
                if abs(word.top - line.avg_line_top) > 5:
                    sus_word_idx.append(i)

            for idx in sus_word_idx:
                tested_word = line.words.pop(idx)
                if not line.is_position_occupied(tested_word.left):
                    line.add_word(tested_word)
                    

    def _extract_order_items(self):
        top_limit, bottom_limit = self._get_order_items_area()

        for line in self._lines:
            # check if line is in orders area
            if line.line_top < top_limit or line.line_bottom > bottom_limit: continue
            
            # check is line is not indented
            if line.line_start > LINE_START_LIMIT: continue
            
            # disregard subtotal lines
            # TODO: make this more robust
            if "reszosszeg" in line.text.lower(): continue

            # handling discount lines
            if "e" in line.words[-1].text or "kedvezmeny" in line.text.lower() or "akci" in line.text.lower():
                self._apply_discount_on_item(line)
                continue
            
            # at this point we consider the line as order line
            item_name, item_price = line.extract_item_data()
            if item_name is None: continue
                

            item = Item(line.line_num, item_name, item_price, 0, line.text)
            self.order_items.append(item)

    def _extract_total(self) -> int:
        # this sets it to None in case 'fizetend' pattern is not matched
        price = None

        for line in self._lines:
            if "fizetend" in line.text.lower() or "osszesen" in line.text.lower():
                _, price = line.extract_item_data()
        
        return price
    
    def _extract_date(self) -> datetime|None:
        _, bottom_limit = self._get_order_items_area()
        date_string_parts:list[str] = []

        for line in self._lines:
            if line.line_top < bottom_limit: continue

            if "tid" in line.text.lower():
                for word in line.words:
                    if word.left > DATE_MIN_LEFT_BOUNDARY:
                        date_string_parts.append(word.text)

        # Define the format that matches your date strings
        date_format = "%Y.%m.%d %H:%M"
        date_string = " ".join(date_string_parts)

        try:
            # Convert to datetime objects
            date = datetime.strptime(date_string, date_format)
        except:
            date = None
        
        return date

    
    # TODO: it feels not too robust
    def _apply_discount_on_item(self, discount_line:Line):
        _, discount_amount = discount_line.extract_item_data()
        for item in reversed(self.order_items):
            if item.line_num == discount_line.line_num - 1 :
                item.add_discount(-discount_amount)
                break
        
        
    def to_json(self) -> str:
        """
        Returns the object's data in a json string.
        """
        items = [{"name": item.name, "price": item.price, "discount": item.discount,  "parsed_line": item.raw_text} for item in self.order_items]

        data = {
            # TODO: parse dat from the invoice
            "datetime": self.get_datetime_str(),
            "items": items,
            "parsed_total": self.invoice_total,
            "calculated_total": self.calculate_total()
        }

        return json.dumps(data, indent=4)

    def calculate_total(self) -> int:
        return sum([item.price for item in self.order_items])
    
    def get_datetime_str(self) -> str:
        if self.invoice_date is None:
            return "missing datetime"
        
        return self.invoice_date.strftime("%Y-%m-%d %H:%M:%S")

    
if __name__ == "__main__":

    invoice_paths = [
        "16000333862025032623918",  
        "16000335892025032535080",
        "16000333892025031436070",
        "16000335892025032735441",
        "16000333220250508145"
    ]



    for inv_name in invoice_paths:

##### Exploration
        # img = preprocess_image(inv_path)
        # custom_config = r'--psm 6'
        # text_raw:str = pytesseract.image_to_data(img, lang="eng", config=custom_config)
        # # print(text_raw)

        # data_raw:str = pytesseract.image_to_data(img, lang="eng", output_type=pytesseract.Output.DICT, config=custom_config)

        # seen_values = set()
        # filtered_tops = [(top + data_raw["height"][i], data_raw["text"][i]) for i, top in enumerate(data_raw["top"]) if data_raw["text"][i] != ""] 
        # sorted_values = sorted(filtered_tops, key=lambda x: x[0])
        # prev_value = 0

        # for i in range(len(filtered_tops)):
        #     top = sorted_values[i]

        #     if top[0] in seen_values:
        #         continue
                
        #     print(top[0], f"+{top[0] - prev_value}", top[1])
        #     seen_values.add(top[0])
        #     prev_value = top[0]

########## end exploration

        # data_raw:str = pytesseract.image_to_data(inv_path, lang="eng", output_type=pytesseract.Output.DICT ,config=custom_config)

# latest work test
        data_dict = process_invoice(f"python_ocr_extractor/test_invoices/{inv_name}.png")
        # print(data_dict)
        invoice = Invoice(data_dict)

        print(invoice.to_json())

        with open(f"python_ocr_extractor/test_invoices/output/{inv_name}.json", "w") as file:
            file.write(invoice.to_json())

        # invoice._extract_date()
        
        
    

