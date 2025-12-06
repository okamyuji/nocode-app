import { FieldType } from "@/types";
import {
  FiCalendar,
  FiCheckSquare,
  FiFile,
  FiHash,
  FiLink,
  FiList,
  FiType,
} from "react-icons/fi";
import { MdOutlineTextFields, MdRadioButtonChecked } from "react-icons/md";

// フィールドタイプとアイコンのマッピング
export const fieldTypeIcons: Record<FieldType, React.ComponentType> = {
  text: FiType,
  textarea: MdOutlineTextFields,
  number: FiHash,
  date: FiCalendar,
  datetime: FiCalendar,
  select: FiList,
  multiselect: FiList,
  checkbox: FiCheckSquare,
  radio: MdRadioButtonChecked,
  link: FiLink,
  attachment: FiFile,
};
