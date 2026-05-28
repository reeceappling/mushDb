interface TextBoxProps {
    readonly: boolean,
    label: string,
    value: string,
    fieldName: string,
    updateTextHandler: (txt: string) => void,
    numbersOnly?: boolean
    classes?: string
}



export default function TextBox(props: TextBoxProps) {
    const doNumsOnly = props.numbersOnly || false
    const internalUpdateHandler = (s: string)=>{
        let current = s
        if(doNumsOnly){
            current = current.replace(/\D/,'')
        }
        props.updateTextHandler(current)
    }
    return (
        <div className={props.classes || ""}>
            <label htmlFor={props.fieldName}>{props.label}</label>
            <input type={"text"} className={"bg-white"} readOnly={props.readonly} name={props.fieldName} value={props.value} onChange={(e) => {internalUpdateHandler(e.currentTarget.value)}} onRateChangeCapture={(e) => {internalUpdateHandler(e.currentTarget.value)}}/>
        </div>
    )
}
