import {
    AllEntries,
    AreaProps,
    Data,
    FormListArea,
    GroupProps,
    SplitAllEntries
} from "@/app/components/formSubcomponents/shared";
import NotesArea, {NotesAreaOld, IsValidNote, Note, NotesGrid} from "@/app/components/formSubcomponents/notes";
import {ChangeEvent} from "react";
import DateArea from "@/app/components/formSubcomponents/date";
import {CheckArrayType} from "@/app/components/common";
import {ExamplePicsWithNotesIncoming} from "@/app/components/formSubcomponents/contaminations";

export interface PicWithNotesNew {
    time: number
    location: string,
    notes?: Note[],
}

export interface PicWithNotesForm {
    time: number
    img: string,
    notes: AllEntries<Note>,
}

export interface NewPicWithNotesForm { // TODO: USE THIS
    time: number
    img: File | undefined,
    notes: AllEntries<Note>,
}

export function InitialPicsEntries(initialPics?: PicWithNotesIncoming[]): SplitAllEntries<PicWithNotesForm,NewPicWithNotesForm>{
    let initialEntries: Data<PicWithNotesForm>[] = initialPics===undefined?[]:initialPics.map((p)=>{
        let ns: Note[] = p.notes || []
        return {
            data:{
                time:p.time,
                img:p.location,
                notes:{existing:ns.map((n)=>{
                        return {data: n,disabled: false}
                    }),new:[]},
            },
            disabled: false,
        }
    })
    return {existing:initialEntries,new:[]}
}

const PicsEndpoint = "/db/images/" // TODO: ENSURE THIS WORKS ALONGSIDE THE main.go ONE

export function ImageLocationFor(str: string) {

    return PicsEndpoint+str
}

export default function PicWithNotesArea(props: AreaProps<PicWithNotesForm>){ // TODO: NEED TO USE imageSelector for new items!!!!!!
    return FormListArea(PicWithNotesEntriesGroup)(props)
}

function PicWithNotesEntriesGroup({initialEntries, preexisting, readonly, updateParent}: GroupProps<PicWithNotesForm>){
    const handleImageChange = (index: number) => { // TODO: FIXME!!!!!
        return function(e: ChangeEvent<HTMLInputElement>){
            let data = [...(initialEntries || [])];
            let files = e.currentTarget.files
            if(files === null || files.length<1){
                return
            }
            var reader = new FileReader();
            reader.onloadend = function() {
                if (typeof reader.result === "string") {
                    data[index].data.img = reader.result;
                    updateParent(data)
                } else {
                    // TODO: SOMETHING HERE?
                }
            }
            // TODO: https://stackoverflow.com/questions/12368910/html-display-image-after-selecting-filename
        }
    }
    const handleNotesChange = (index: number) => {
        return function(newNotes: AllEntries<Note>) {
            let data = [...(initialEntries || [])];
            data[index].data.notes = newNotes
            updateParent(data)
        }
    }
    const handleTimeChange = (index: number) => {
        return function(newTime: number) {
            let data = [...(initialEntries || [])];
            data[index].data.time = newTime
            updateParent(data)
        }
    }
    const addFields = () => {
        let data = [...(initialEntries || []), { data: {notes: {existing:[], new:[]}, img: "", time: Date.now()}, disabled: false }]
        updateParent(data)
    }
    const removeFields = (index: number) => {
        return () => {
            let data = [...(initialEntries || [])];
            data.splice(index, 1) // TODO: THIS WONT WORK PROPERLY WITH INDEX
            updateParent(data)
        }
    }
    const disableField = (index: number) => {
        return () => {
            let data = [...(initialEntries || [])]
            data[index].disabled = !data[index].disabled
            updateParent(data)
        }
    }
    const groupClasses = () => {
        let out = ""
        if(preexisting){
            out = "exists"
        } else {
            out = "new"
        }
        if(readonly){
            out+=" readonly"
        } else {
            out+=" editable"
        }
        return out
    }
    const entryClasses = (note: Data<PicWithNotesForm>) => {
        let out = "picWithNotes"
        if(note.disabled){
            out+=" disabled"
        } else {
            out+=" enabled"
        }
        return out
    }
    return (<div className={groupClasses()}>{/* TODO: CLASS STYLINGS!!!! */}
        {initialEntries?.map((input, index) => {
            return (
                <div className={entryClasses(input)} key={index}>
                    <DateArea when={initialEntries[index].data.time} readonly={readonly} updateParent={handleTimeChange}/>{/* TODO: date?*/}
                    {input.disabled?"disabled":null /* TODO: remove? */}
                    {/* TODO: INPUT TAG */}
                    {input.data.img != "" ?
                        <img src={preexisting?ImageLocationFor(input.data.img):input.data.img}/> /* TODO: NO imageLocationFor in case of new images */:
                        <input type={"file"} accept={"image/*;capture=camera"} id={"file_field_2"} name={"file_field_1"} capture={"user"} onInput={(e)=>{handleImageChange(index)}}/>
                    }
                    {/* TODO: INPUT TAG. NotesFormArea? */}
                    <NotesGrid current={input.data.notes} readonly={readonly} updateParent={handleNotesChange(index)}/>{/* TODO: ensure handleNotesChange is ok*/}
                    {readonly ? null : <button onClick={()=>{preexisting?disableField(index):removeFields(index)}}>{preexisting ? (input.disabled?"enable":"disable") : "remove"}</button>}
                </div>)
        })}
        {preexisting ? null : <button className={"basicButton"} onClick={addFields}>Add More..</button>}
    </div>)
}

// function PicWithNotesEntriesGroupOld({initialEntries, preexisting, readonly, updateParent}: GroupProps<PicWithNotesForm>){
//     const [inputFields, setInputFields] = useState(initialEntries)
//     const handleImageChange = (index: number) => { // TODO: FIXME!!!!!
//         return function(e: ChangeEvent<HTMLInputElement>){
//             let data = [...(inputFields || [])];
//             let files = e.currentTarget.files
//             if(files === null || files.length<1){
//                 return
//             }
//             var reader = new FileReader();
//             reader.onloadend = function() {
//                 if (typeof reader.result === "string") {
//                     data[index].data.img = reader.result;
//                     updateParent(data)
//                     setInputFields(data);
//                 } else {
//                     // TODO: SOMETHING HERE?
//                 }
//             }
//             // TODO: https://stackoverflow.com/questions/12368910/html-display-image-after-selecting-filename
//         }
//     }
//     const handleNotesChange = (index: number) => {
//         return function(newNotes: AllEntries<Note>) {
//             let data = [...(inputFields || [])];
//             data[index].data.notes = newNotes
//             updateParent(data)
//             setInputFields(data);
//         }
//     }
//     const handleTimeChange = (index: number) => {
//         return function(newTime: number) {
//             let data = [...(inputFields || [])];
//             data[index].data.time = newTime
//             updateParent(data)
//             setInputFields(data);
//         }
//     }
//     const addFields = (e: MouseEvent) => {
//         e.preventDefault()
//         let data = [...(inputFields || []), { data: {notes: {existing:[], new:[]}, img: "", time: Date.now()}, disabled: false }]
//         updateParent(data)
//         setInputFields(data)
//     }
//     const removeFields = (index: number) => {
//         return (event: MouseEvent) => {
//             event.preventDefault()
//             let data = [...(inputFields || [])];
//             data.splice(index, 1) // TODO: THIS WONT WORK PROPERLY WITH INDEX
//             updateParent(data)
//             setInputFields(data)
//         }
//     }
//     const disableField = (index: number) => {
//         return (event: MouseEvent) => {
//             event.preventDefault()
//             let data = [...(inputFields || [])]
//             data[index].disabled = !data[index].disabled
//             updateParent(data)
//             setInputFields(data)
//         }
//     }
//     const groupClasses = () => {
//         let out = ""
//         if(preexisting){
//             out = "exists"
//         } else {
//             out = "new"
//         }
//         if(readonly){
//             out+=" readonly"
//         } else {
//             out+=" editable"
//         }
//         return out
//     }
//     const entryClasses = (note: Data<PicWithNotesForm>) => {
//         let out = "picWithNotes"
//         if(note.disabled){
//             out+=" disabled"
//         } else {
//             out+=" enabled"
//         }
//         return out
//     }
//     return (<div className={groupClasses()}>{/* TODO: CLASS STYLINGS!!!! */}
//         {inputFields?.map((input, index) => {
//             return (
//                 <div className={entryClasses(input)} key={index}>
//                     <DateArea when={inputFields[index].data.time} readonly={readonly} updateParent={handleTimeChange}/>{/* TODO: date?*/}
//                     {input.disabled?"disabled":null /* TODO: remove? */}
//                     {/* TODO: INPUT TAG */}
//                     {input.data.img != "" ?
//                         <img src={preexisting?ImageLocationFor(input.data.img):input.data.img}/> /* TODO: NO imageLocationFor in case of new images */:
//                         <input type={"file"} accept={"image/*;capture=camera"} id={"file_field_2"} name={"file_field_1"} capture={"user"} onInput={(e)=>{handleImageChange(index)}}/>
//                     }
//                     {/* TODO: INPUT TAG */}
//                     <NotesArea current={input.data.notes} readonly={readonly} updateParent={handleNotesChange(index)}/>{/* TODO: ensure handleNotesChange is ok*/}
//                     {readonly ? null : <button onClick={()=>{preexisting?disableField(index):removeFields(index)}}>{preexisting ? (input.disabled?"enable":"disable") : "remove"}</button>}
//                 </div>)
//         })}
//         {preexisting ? null : <button className={"basicButton"} onClick={()=>{addFields}}>Add More..</button>}
//     </div>)
// }

// TODO: SINGLE PicWithNotes!!!!!

export interface PicWithNotesIncoming {
    time: number
    location: string
    notes?: Note[]
}

export function IsValidPicWithNotesIncoming(input: any): boolean {
    return (
        typeof input === 'object' &&
        'time' in input && typeof input.time === 'number' &&
        'location' in input && typeof input.location === 'string' &&
        (('notes' in input)?Array.isArray(input.notes)&&CheckArrayType(input.notes,IsValidNote):true)
    )
}
export const ExamplePicWithNotesIncoming: PicWithNotesIncoming = ExamplePicsWithNotesIncoming[0]