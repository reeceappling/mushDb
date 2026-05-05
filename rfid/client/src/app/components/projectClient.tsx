'use client'

import React, {ChangeEvent, useEffect, useState} from "react";
import NotesAreaOld, {
    IsValidNote, NewEntryNotes,
    Note,
    NoteEntriesGroup,
    NotesAreaInline
} from "@/app/components/formSubcomponents/notes";
import {AllEntries, Data} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea, {NumberToDate} from "@/app/components/formSubcomponents/date";
import {
    DisplayInput,
    HandleJsonResponse,
    HeaderLevel,
    InlineExpansionArea, InlineExpansionButton,
    InlineProps,
    InlineSubArea, ListPageItems, NewEntryInput,
    OptionalArrayOfType, OptionalKey,
    OptionalSimpleKey, RequiredKey
} from "@/app/components/common";
import {redirect} from "next/navigation";
import {ErrorDisplay} from "@/app/components/formSubcomponents/commonClient";
import {ProjectData, ProjectSelector,} from "@/app/components/projectServer";
import {BaseExternalUrl} from "@/app/components/Constants";
import EntryLink from "@/app/components/formSubcomponents/entryLink";
import {ProjectWithPerm} from "@/app/components/perms";
import {useCookies} from "react-cookie";
import {dataFor} from "@/app/components/agarRecipeClient";
import {SelectorFor, SelectorResetsOnSelectFor} from "@/app/components/selector";
import {IsStringMapToBool, IsStringMapToString} from "@/app/components/accessControlClient";
import {HandleErr, UserSelector} from "@/app/components/userClient";
import TestAndValidate from "@/app/components/testing/untested";
import {DisplayFormWrapper, NewEntryFormWrapper} from "@/app/components/lcRecipeClient";
import {
    FlexedArea,
    FlexedSinglesGroup, ListPageTable,
    ListTableColumn,
    NewColumn,
    NotesFormArea, NumberToDateStr
} from "@/app/components/agarBatchClient";
import {InitialNotesState} from "@/app/components/formSubcomponents/contaminations";
import {TailwindButton} from "@/app/components/tailwind/components";
import {SaleData} from "@/app/components/saleServer";
// TODO: list page not working
// TODO: ensure display page doing what we want

export function AssertProject(input: any): asserts input is ProjectData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    let requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['creationDate', 'number'],
        ['lastUpdated', 'number'],
    ])
    for (let [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Project assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }

    // optional simple keys
    let optionalSimpleKeys = new Map<string, string>([
        ['completed', 'number'],
    ])
    for (let [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('Project assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex required keys
    let complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['perms', IsStringMapToString], // TODO: THIS IS NOT WORKING PROPERLY!!! CHANGE TO STRING FORMAT!!!!
    ])
    for (let [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('Project assertion failure: required key ' + key + ' was not valid. was '+JSON.stringify(input[key]));
        }
    }
    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['notes', IsValidNote],
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Plate assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    return
}

export default function ProjectDisplay(
    {
        id, readonly, data, headerLevel, cookies
    }: DisplayInput) {
    try {
        AssertProject(data)
        const [initial, setInitial] = useState(data)
        const initPerms =  new Map<string, string>(Object.entries(data.perms||{}) as [string, string][]) // TODO: IF THIS WORKS USE IT FOR UNMARSHALLING ALL PARMS!

        const [completed, setCompleted] = useState(data.completed)
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
        const [perms, setPerms] = useState<Map<string, string>>(initPerms)
        const [err, setErr] = useState<string | undefined>()
        const updateInitial = (updated: ProjectData)=>{
            setInitial(updated)
            const ps =  new Map<string, string>(Object.entries(updated.perms||{}) as [string, string][]) // TODO: IF THIS WORKS USE IT FOR UNMARSHALLING ALL PARMS!

            setCompleted(updated.completed)
            setNotes(InitialNotesState(updated.notes))
            setPerms(ps)
        }
        // TODO: users and entries for this!
        const completedArea = () => {
            const handleCompletedClick = () => {
                setCompleted(completed ? undefined : Date.now())
            }
            if (readonly || initial.completed) {
                let isComp = "In-Progress"
                if (completed) {
                    isComp = "Completed " + NumberToDate(new Date(completed))
                }
                return <div>
                    <div>{isComp}</div>
                </div>
            }
            return <div>
                <div>{"Completed: "}</div>
                <input type={"checkbox"} checked={!!completed} onChange={()=>{
                    // TODO: ensure onChange does not need anything
                }} onClick={handleCompletedClick} onSubmit={(e)=>{e.preventDefault();}}/>
            </div>
        }
        const projectSubmit = () => {
            let body: any = {
                notes: notes,
                completed: completed,
                perms: Object.fromEntries(perms), // TODO: ensure this is being done on any maps that are being marshalled!!!!!
            }
            console.log("sending perms: "+JSON.stringify(Object.fromEntries(perms)))


            fetch(BaseExternalUrl + "/db/update/project/"+encodeURIComponent(data._id), { // TODO: question marks in id cause issues
                method: "POST",
                headers: {
                    credentials: 'include',
                     'Cookie': cookies,
                    'Content-type': "application/json"
                },
                body: JSON.stringify(body)
            }).then(HandleJsonResponse)
                .then((entry) => {
                    AssertProject(entry)
                    updateInitial(entry)
                })
                .catch((err) => {
                    HandleErr(err,setErr)
                });
        }
        return (
            <DisplayFormWrapper entryType={"project"}>
                <ErrorDisplay err={err} headerLevel={headerLevel}/>
                {/* data._id on next line because project name can have spaces?*/}
                <ID id={data._id} txt={"Project"} entryType={"project"}/>
                <FlexedArea>
                    <FlexedSinglesGroup>
                        <DateArea pre={"Created: "} when={initial.creationDate} readonly={true}/>
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        {completedArea()}
                    </FlexedSinglesGroup>
                </FlexedArea>
                <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
                <TestAndValidate todos={[
                    "setting a user to view only and updating will remove the user from the project :(",
                "allow users to be removed!"]}>{/* TODO: THIS*/}
                    <ProjectPermsArea perms={perms} setPerms={setPerms} readonly={readonly}/> {/* TODO: HEAVILY TEST! Also ensure this is properly covered on the go side!*/}
                </TestAndValidate>
                {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e)=>{
                    e.stopPropagation();
                    projectSubmit()
                }}>{"Update"}</button>}
            </DisplayFormWrapper>
        )
    } catch (err) {
        return <div>{"ERROR: Plate data format incorrect: " + err}</div>
    }
}

export function NewProjectForm(
    {handlers}: {handlers: NewEntryInput<ProjectData>}) { // TODO: add cookies?
    const [name, setName] = useState<string | undefined>(undefined)
    const [notes, setNotes] = useState<Note[]>([])
    // TODO: load up user on server side into the userperms as write (unless blanket is write)
    const [err, setErr] = useState<string | undefined>(undefined)
    // TODO: handle isTopLevel
    const createProject = () => {
        if (name === undefined) {
            setErr("Name field cannot be undefined")
            return
        }
        let body = {
            name: name, // TODO: validate that project name is valid for url
            notes: notes,
        }
        fetch(BaseExternalUrl + "/db/create/project", {
            method: "POST",
            headers: {
                credentials: 'include',
                'Content-type': 'application/json'
            },
            body: JSON.stringify(body)
        })
            .then(HandleJsonResponse)
            .then((proj) => {
                AssertProject(proj)
                handlers.onCreate && handlers.onCreate(proj)
            })
            .catch((error) => {
                HandleErr(error, setErr)
            });
    }
    const handleChangeName = (event: ChangeEvent<HTMLInputElement>) => {
        setName(event.target.value)
    }
    const updateNotes = (entries: AllEntries<Note>) => {
        setNotes(entries.new.map((e) => {
            return e.data
        }))
    }
    const projectNameArea = () => {
        return <div>
            <div>{"Project Name: "}</div>
            <input type={"text"} value={name || ""} onChange={handleChangeName}/>
        </div>
    }
    return <NewEntryFormWrapper entryType={"project"}>
        <ErrorDisplay err={err}/>
        {projectNameArea()}
        {/* TODO TODO: ensure notes delete themselves when disabled*/}
        <NewEntryNotes setNotes={setNotes}/>
        <TailwindButton txt={"Create"} click={createProject}/>
    </NewEntryFormWrapper>

}

// export function ProjectInline({data, expandByDefault, onClick, showMainPageButton, idIsLink}: InlineProps<ProjectData>) {
//     const [expanded, setExpanded] = useState(expandByDefault)
//     return <div>
//         <InlineSubArea props={{}}>
//             <ID id={data._id} txt={"Project"} entryType={"project"} allowOpenMainPage={showMainPageButton} linkPage={idIsLink}/>
//             <DateArea pre={"Created: "} when={data.creationDate} readonly={true}/>
//             {data.completed ? <div>
//                 <div>{"Completed: " + NumberToDate(new Date(data.completed))}</div>
//             </div> : <div>
//                 <div>{"In-Progress"}</div>
//             </div>}
//         </InlineSubArea>
//         <InlineExpansionArea props={{expanded: expanded}}>
//             <NotesAreaInline notes={data.notes} headerLevel={headerLevel} offset={-1}/>
//             <DateArea pre={"Last Updated: "} when={data.lastUpdated} readonly={true}/>
//         </InlineExpansionArea><InlineExpansionButton data-cy-id="InlineSubAreaButton" setExpanded={setExpanded}
//                                expanded={expanded}/>
//     </div>
// }

// export function ProjectsArea( // TODO: is this even needed anymore?
//     {projects, allowCreate, allowRemove, headerLevel, readonly, offset, setProjects}: {
//         projects?: string[],
//         allowCreate: boolean
//         allowRemove: boolean
//         headerLevel?: number,
//         readonly?: boolean
//         setProjects?: (ps: string[]) => void
//         offset?: number
//     }) {
//     const [current, setCurrent] = useState<Data<string>[]>((projects || []).map(p => {
//         return {data: p, disabled: false}
//     }))
//     const [creatorOpen, setCreatorOpen] = useState(false)
//     const updateEntries = (newProjs: Data<string>[]) => {
//         let upd = newProjs.filter(s => !s.disabled).map(v => {
//             return v.data
//         })
//         setProjects && setProjects(upd)
//         setCurrent(newProjs)
//     }
//     const creationArea = () => {
//         return <div>
//             <div className={"centerH"}>
//                 <button className={"gapTop"} onClick={() => {
//                     setCreatorOpen(!creatorOpen)
//                 }}>{creatorOpen ? "Close Project Selector" : "Add a project"}</button>
//             </div>
//             {creatorOpen && <ProjectSelector doSelect={proj => {
//                 if (proj === undefined) {
//                     return
//                 }
//                 // TODO: if project already exists on list, dont add!
//                 let newProj = {data: proj._id, disabled: false}
//                 let out = [...current]
//                 out = [...out, newProj]
//                 proj && updateEntries(out)
//                 setCreatorOpen(false)
//             }} headerLevel={headerLevel} creatorInPage={false} allowCreation={true}/>}
//         </div>
//     }
//     const ap: React.ReactNode = <div className={"centerH"}>{"Associated Projects: "}</div> // TODO: OK?
//     const ps = () => {
//         if (current.length === 0) {
//             return <div>{"None"}</div>
//         }
//         return <div className={"projectsFlex"}>{current.map((proj, i) => {
//             if (proj.disabled) {
//                 return null
//             }
//             const projLinkName = "WITH UNDERSCORES" // TODO: THIS
//             return <div key={i} className={"gridCol"}>
//                 <EntryLink
//                     props={{displayedId: proj.data, linkId: projLinkName, entryType: "project", openInNewTab: true}}>
//                     <div>{proj.data}</div>
//                     {/* TODO: TEXT SIZE*/}
//                 </EntryLink>
//                 {(allowRemove && !readonly) && <button onClick={() => {
//                     let out = [...current]
//                     out[i].disabled = true
//
//                     updateEntries(out)
//                 }}>{"Remove"}</button>}
//             </div>
//         })}</div>
//     }
//
//     return <div>
//         {ap}
//         {ps()}
//         {(!readonly && allowCreate) && creationArea()}
//     </div>
// }

export function NumberToPerm(n?: number) {
    if (n === undefined || n === 0) {
        return undefined
    }
    return n === 2;
}

export function PermToNumber(p: boolean | undefined): number {
    if (p === undefined) {
        return 0
    }
    if (!p) {
        return 1
    }
    return 2
}

export function ReadWriteAdminSelector({readonly,onUpdate, value}:{value:string,readonly:boolean,onUpdate?:(valOut: string)=>void}){
    const strForVal = (str?: string)=>{
        return (value==="read")?"can view":(value==="admin"?"is admin":"can edit")
    }
    if (readonly) {
        return <text>{strForVal(value)}</text>
    }
    return <SelectorFor options={["can view", "can edit", "is admin"]} initial={strForVal(value)}
                        updateParent={s => {
                            if (s === "is admin") {
                                onUpdate && onUpdate("admin")
                            }
                            if (s === "can edit") {
                                onUpdate && onUpdate("write")
                            }
                            if (s === "can view") {
                                onUpdate && onUpdate("read")
                            }
                            return
                        }} disabled={false}/>
}

export function ReadWriteSelector({readonly,onUpdate, value}:{value:boolean,readonly:boolean,onUpdate?:(b: boolean)=>void}){
    if (readonly) {
        return <text>{value?"can edit":"can view"}</text>
    }
    return <SelectorFor options={["can view", "can edit"]} initial={value ? "can edit" : "can view"}
                        updateParent={s => {
                            onUpdate && onUpdate(s==="can edit")
                        }} disabled={false}/>
}

export function ProjectPermsArea({perms, setPerms, readonly}: {
    perms?: Map<string, string>, // TODO: to string!
    setPerms?: (pp: Map<string, string>) => void, // TODO: to string!
    readonly: boolean,
}) {
    if(readonly && !perms){
        return null
    }
    return <div><TestAndValidate todos={["title area"]}>
        {perms!==undefined && perms.size>0 && [...perms.entries()].map((p) => { // TODO: how to handle the undefineds?
            return <div key={p[0]}>
                {p[0]}
                <ReadWriteAdminSelector readonly={readonly} value={p[1]} onUpdate={(b) => {
                    setPerms && setPerms(new Map(perms).set(p[0], b))
                }}/>
            </div>
        })}
        {/* AREA TO ADD USER */}
        <UserSelector onSelect={(u)=>{
            let out = (perms !==undefined && perms.size>0)?new Map<string, string>(perms):new Map<string, string>()
            setPerms && setPerms(out.set(u._id, "read"))
        }} blacklist={(perms!==undefined&&perms.size>0)?[...perms.entries()].map(u=>{return u[0]}):[]}/>
    </TestAndValidate>
    </div>
}

// // TODO: FIX FOR NEW PROJECT PERMS!
// export function ProjectPermsAreaOLD({originalPerms, setPerms, canEdit}: { // TODO: this whole thing?
//     originalPerms?: ProjectPerms,
//     setPerms?: (pp: ProjectPerms) => void,
//     canEdit: boolean,
// }) {
//     const [fullPerms, setFullPerms] = useState<ProjectPerms>(originalPerms || {
//         users: {ids: [], canWrite: []},
//         blanket: 0
//     })
//     const [userToAdd, setUserToAdd] = useState<string>("")
//     const [userToAddCanWrite, setUserToAddCanWrite] = useState<boolean>(false)
//     const [err, setErr] = useState<string | undefined>()
//     //const [cookies, setCookie, removeCookie] = useCookies(['SessionId']);
//     if (!canEdit) {
//         const blanketText = () => {
//             switch (fullPerms.blanket) {
//                 case 0:
//                     return ""
//                 case 1:
//                     return "Read"
//                 case 2:
//                     return "Read, Write"
//             }
//             if (fullPerms.blanket === 0) {
//                 return "No blanket permissions"
//             }
//         }
//         const textForCanWrite = (canWrite: boolean) => {
//             if (canWrite) {
//                 return "Write"
//             }
//             return "Read"
//         }
//         return <div>
//             <div>{"Blanket permissions: " + blanketText()}</div>
//             <div>
//                 <div>{"User permissions"}</div>
//                 <table>
//                     <tr>
//                         <th>{"User"}</th>
//                         <th>{"Permission"}</th>
//                     </tr>
//                     {fullPerms.users.ids.map((user, i) => {
//                         return <tr key={i}>
//                             <td>{user.val}</td>
//                             <td>{textForCanWrite(fullPerms.users.canWrite[i])}</td>
//                         </tr>
//                     })}
//                 </table>
//             </div>
//         </div>
//     }
//
//
//     const doUpdate = (newPerms?: ProjectPerms) => {
//         let tempPerms = newPerms || {users: {ids: [], canWrite: []}, blanket: 0}
//         if (newPerms) {
//             tempPerms = newPerms
//         }
//         setFullPerms(tempPerms)
//         if (!setPerms) {
//             return
//         }
//         // remove unnecessary perms as needed
//         if (tempPerms.blanket === 0) {
//             setPerms(tempPerms)
//             return
//         }
//         let toSet: ProjectPerms = {users: {ids: [], canWrite: []}, blanket: tempPerms.blanket}
//         if (tempPerms.blanket === 1) {
//             for (let i = 0; i < tempPerms.users.ids.length; i++) {
//                 if (tempPerms.users.canWrite[i]) {
//                     toSet.users.ids.push(tempPerms.users.ids[i])
//                     toSet.users.canWrite.push(true)
//                 }
//             }
//         }
//         setPerms(toSet)
//     }
//
//     const changeBlanket = (b?: boolean) => {
//         let updated = {...fullPerms}
//         updated.blanket = PermToNumber(b)
//         doUpdate(updated)
//     }
//     const changeUser = (un: number) => {
//         return (newBp?: boolean) => {
//             let updated = {...fullPerms}
//             if (newBp === undefined) {
//                 // remove that user
//                 const filterFunc = (_: any, index: number) => {
//                     return index !== un
//                 }
//                 updated.users.ids = [...fullPerms.users.ids].filter(filterFunc);
//                 updated.users.canWrite = [...fullPerms.users.canWrite].filter(filterFunc);
//
//             } else {
//                 // Update that user
//                 updated.users.canWrite[un] = newBp
//             }
//             doUpdate(updated)
//         }
//     }
//     const addUserByEmailOrUsername = (val: string) => {
//         if (setPerms) {
//             fetch(BaseExternalUrl + "/db/userIdFor", { // TODO: handle on the go side
//                 method: 'Get',
//                 body: val,
//                 headers: {
//                     credentials: 'include',
//                     'Cookie': cookies,
//                     //'Content-type': "application/json"
//                     //Authorization: tokenFetch, // TODO: auth?
//                 },
//             }).then(HandleTxtResponse).then((id) => {
//                 let updated = {...fullPerms}
//                 updated.users.ids.push({id: id, val: val})
//                 updated.users.canWrite.push(userToAddCanWrite)
//                 doUpdate(updated)
//             }).catch((e: string) => {
//                 setErr(e)
//             })
//         }
//     }
//     return <div>
//         <div>
//             {"Global Permission: "}<PermissionSelector onChange={changeBlanket}
//                                                        canWrite={NumberToPerm(fullPerms.blanket)}/>
//         </div>
//         {/* TODO: REMOVE ALL UNNECESSARY USERS IF BLANKET CHANGE SHOULD*/}
//         <div>
//             <div>{/* Current Users */}
//                 {fullPerms.users.ids.map((uid: UserIdPair, i: number) => {
//                     return <div className={"projPermsAreaUser"} key={i}>
//                         {uid.val + ": "}<PermissionSelector onChange={changeUser(i)}
//                                                             canWrite={fullPerms.users.canWrite[i]}/>
//                     </div>
//                 })}
//             </div>
//             <div>
//                 <ErrorDisplay err={err}/>
//                 {/* Add new user by email */}
//                 <Textbox readonly={false} label="New user email or username" value={userToAdd} fieldName={"FIXME"}
//                          updateTextHandler={setUserToAdd}/>
//                 <PermissionSelector onChange={b => {
//                     b && setUserToAddCanWrite(b)
//                 }} canWrite={userToAddCanWrite}
//                                     dontShowBelow={fullPerms.blanket === 0 ? false : NumberToPerm(fullPerms.blanket)}/>
//                 <button onClick={() => {
//                     addUserByEmailOrUsername(userToAdd)
//                 }}> {"ADD USER"}</button>
//             </div>
//         </div>
//         <button onClick={() => {
//             doUpdate(originalPerms || {users: {ids: [], canWrite: []}, blanket: 0})
//         }}>{"Revert permission changes"}</button>
//     </div>
// }

export async function GetMyProjects() {
    return await fetch(BaseExternalUrl + "/sessionUserProjects", { // TODO: ensure works like we want! We JUST want the user's perms on each project
        method: "GET",
        headers: {
            credentials: 'include',
            // TODO: SessionId: sessionId,
            'Content-type': "application/json"
        }
    }).then(HandleJsonResponse).then((projs) => {
        try {
            return projs as ProjectWithPerm[]
        } catch (err) {
            throw err
        }
    })
}

export async function GetSessionUserProjects(sessionId: string): Promise<ProjectWithPerm[]> {
    return fetch(BaseExternalUrl + "/sessionUserProjects", { // TODO: ensure works like we want! We JUST want the user's perms on each project
        method: "GET",
        headers: {
            credentials: 'include',
            SessionId: sessionId,
            'Content-type': "application/json"
        }
    }).then(HandleJsonResponse).then((projs) => {
        try {
            return projs as ProjectWithPerm[]
        } catch (err) {
            throw err
        }
    })
}

export function ProjectsSelector(inp: {
    onSelect: (projName: string) => void
    blacklist?: string[]
}) {
    const [loading, setLoading] = useState(true)
    const [projects, setProjects] = useState<string[]>([])
    const [err, setErr] = useState<string | undefined>(undefined)
    // TODO: does this need incremented depth?
    useEffect(() => {
            fetch(BaseExternalUrl + "/sessionUserProjects", { // TODO: do we also want to pull the user's perms for each project?
                method: "GET",
                headers: {
                    credentials: 'include',
                    // TODO: SessionId: sessionId,
                    'Content-type': "application/json"
                }
            }).then(HandleJsonResponse).then((projs) => {
                try {
                    return projs as string[] // TODO: FIXME!
                } catch (err) {
                    throw err
                }
            }).then(projs => {
                setProjects(projs)
                setLoading(false)
                setErr(undefined)
                return
            }).catch(err => {
                HandleErr(err, setErr)
                return
            })
        }, [])
    if (loading) {
        return <div>{"Loading projects selector"}</div>
    }
    const projectOptions = ()=>{
       if (projects == undefined || projects.length == 0) {
           return [""]
       }
        return ["", ...projects.filter(pToFilter=>{
                return (inp.blacklist || []).indexOf(pToFilter) == -1
            })]
    }
    return <div>
        <ErrorDisplay err={err}/>
        <SelectorResetsOnSelectFor options={projectOptions()} updateParent={(pr)=>{
            console.log("selected "+pr)
            inp.onSelect(pr)
        }}/>
    </div>
}

export function ProjectListPageTable({data, onClick}: ListPageItems<ProjectData>) {
    const cols: ListTableColumn<ProjectData>[] = [
        NewColumn("Name", (v)=>v._id),
        NewColumn("Completed", (v)=>{
            return v.completed?NumberToDateStr(v.completed):""
        }),
        NewColumn("Created", (v)=>{
            return NumberToDateStr(v.creationDate)
        }),
        NewColumn("Updated", (v)=>{
            return NumberToDateStr(v.lastUpdated)
        }),
    ]
    // TODO: expansion for everything else????
    return <ListPageTable cols={cols} data={data} onClick={onClick}/>
}