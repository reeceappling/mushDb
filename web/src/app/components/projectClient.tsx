'use client'

import React, { JSX, useContext, useEffect, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import {AllEntries} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea, {NumberToDate} from "@/app/components/formSubcomponents/date";
import {
    clientPostRequestHeaders,
    DisplayFormWrapper,
    DisplayInput, DoCreateRequest, DoUpdateRequest, ExistingRecentSelector, FlexedArea, FlexedSinglesGroup,
    HandleJsonResponse,
    ListPageItems, ListPageTable, ListTableColumn, NewColumn, NewEntryFormWrapper,
    NewEntryInput, NumberToDateStr,
    OptionalArrayOfType,
    OptionalSimpleKey
} from "@/app/components/common";
import {ErrorDisplay, RemoveButton} from "@/app/components/formSubcomponents/commonClient";
import {ProjectData,} from "@/app/components/projectServer";
import {BaseExternalUrl} from "@/app/components/Constants";
import {ProjectWithPerm} from "@/app/components/perms";
import {SelectorFor, SelectorResetsOnSelectFor} from "@/app/components/selector";
import {HandleErr, UserSelector} from "@/app/components/userClient";
import TestAndValidate from "@/app/components/testing/untested";
import {DepthContext, DepthProvider} from "@/app/components/formSubcomponents/depthContext/depth";
import {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {InputTextInlineTitle} from "@/app/components/formSubcomponents/numericInput";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";

function requireObject(inp: any){
    const typ = typeof inp
    if (typ !== 'object') {
        throw new Error('Input is not an object! Input is ' + typ);
    }
}

export function AssertProject(input: any): asserts input is ProjectData {
    requireObject(input)
    // required simple keys
    const requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['creationDate', 'number'],
        ['lastUpdated', 'number'],
    ])
    for (const [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Project assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }

    // optional simple keys
    const optionalSimpleKeys = new Map<string, string>([
        ['completed', 'number'],
    ])
    for (const [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('Project assertion failure: optional key ' + key + ' was not valid');
        }
    }

    // complex optional keys
    // const complexOptionalKeys = new Map<string, (v: any) => boolean>([
    //     //['perms', IsStringMapToString], // TODO: THIS IS NOT WORKING PROPERLY!!! CHANGE TO STRING FORMAT!!!! (try above)
    // ])
    // for (const [key, validator] of complexOptionalKeys) {
    //     if (!OptionalKey(key, input, validator)) {
    //         throw new Error('Project assertion failure: required key ' + key + ' was not valid. was ' + JSON.stringify(input[key]));
    //     }
    // }
    // complex optional array keys
    const complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['notes', IsValidNote],
    ])
    for (const [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Project assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    // Complex optional keys (perms)
    //console.log("inp perms: "+JSON.stringify(Object.entries(input.perms))) // TODO: del
    input.perms = UnmarshalProjectPermsField(input)
    //console.log("out perms: "+JSON.stringify(Object.entries(input.perms))) // TODO: del
    return
}

export function UnmarshalProjectPermsField(input: any): Map<string,string> {
   // TODO: needs to be able to throw!
   //  if ((!('perms' in input)) || input.perms === undefined) {
   //      // Does not exist or is undefined, return an empty map
   //      return new Map<string, string>()
   //  }
    if ('perms' in input && (input.perms !== undefined)) { // TODO: ensure works properly, we don't want perms getting messed up
        const field = input.perms
        if (typeof field !== 'object' || field === null) {
            throw 'perms field must be a non-null object'
        }
        // Exists, return populated map
        return new Map<string, string>(Object.entries(field as Record<string, string>))
    }
    // Does not exist or is undefined, return an empty map
    return new Map<string, string>()
}

export default function ProjectDisplay(
    {
        id, readonly, data, headerLevel
    }: DisplayInput<ProjectData>) {
        const [initial, setInitial] = useState(data)
        const permsObjAsMap = (inp?: Map<string, string>):Map<string, string>=>{
            if (inp===undefined || inp.size===0){
                return new Map<string, string>();
            }
            return new Map<string, string>(inp.entries().toArray()) // TODO: revert if does not work! HAVE NOT TESTED!
            // return new Map<string, string>(Object.entries(inp ? Object.fromEntries(inp) : {}) as [string, string][])
        }

        const [completed, setCompleted] = useState(data.completed)
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
        const [perms, setPerms] = useState<Map<string, string>>(permsObjAsMap(data.perms) )
        const [err, setErr] = useState<string | undefined>()
        const updateInitial = (updated: ProjectData) => {
            setInitial(updated)

            setCompleted(updated.completed)
            setNotes(InitialNotesState(updated.notes))
            setPerms(permsObjAsMap(updated.perms))
            setErr(undefined)
        }
        // TODO: users and entries for this!
        const completedArea = () => {
            if (readonly || initial.completed) {
                const isComp = completed ? "Completed " + NumberToDate(new Date(completed)) : "In-Progress"
                return <div>
                    <div>{isComp}</div>
                </div>
            }
            return <div>
                <div>{"Completed: "}</div>
                <input type={"checkbox"} checked={!!completed} onChange={() => {
                    // TODO: ensure onChange does not need anything
                }} onClick={e=>{
                    e.stopPropagation();
                    setCompleted(completed ? undefined : Date.now())
                }}/>
            </div>
        }
        const cookies = useContext(CookiesContext)
        const projectSubmit = () => {
            const permsOut =  Object.fromEntries(perms)
            const body: any = {
                notes: notes,
                completed: completed,
                perms: permsOut,
            }
            console.log("sending perms: " + JSON.stringify(permsOut))


            // TODO: Separate project perms request?
            DoUpdateRequest("project",encodeURIComponent(data._id), body, AssertProject, allCookies(cookies))
                .then(v=>{
                    updateInitial(new ProjectData(v))
                })
                .catch(e=>{
                    setErr("failed to update initial: "+JSON.stringify(e))
                })
        }
        return (
            <DisplayFormWrapper entryType={"project"}>
                <ErrorDisplay err={err}/>
                {/* data._id on next line because project name can have spaces?*/}
                <ID props={{id:data._id, txt:"Project", entryType:"project"}}/>
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
                <TestAndValidate todos={["setting a user to view only and updating will remove the user from the project :("]}>{/* TODO: THIS*/}
                    <ProjectPermsArea perms={perms} setPerms={setPerms} /* TODO: ENSURE ADDING/CHANGING USERS MAKES A SEPARATE REQUEST TO THE SERVER?!*/
                                      readonly={readonly}/> {/* TODO: HEAVILY TEST! Also ensure this is properly covered on the go side!*/}
                </TestAndValidate>
                {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e) => {
                    e.stopPropagation();
                    projectSubmit()
                }}>{"Update"}</button>}
            </DisplayFormWrapper>
        )
}

export function NewProjectForm(
    {handlers}: { handlers: NewEntryInput<ProjectData> }) {
    const [name, setName] = useState<string | undefined>(undefined)
    const [notes, setNotes] = useState<Note[]>([])
    const [err, setErr] = useState<string | undefined>(undefined)

    const cookies = useContext(CookiesContext)
    const createProject = () => {
        if (name === undefined) {
            setErr("Name field cannot be undefined")
            return
        }
        const body = {
            name: name, // TODO: validate that project name is valid for url
            notes: notes,
        }
        DoCreateRequest("project", body, AssertProject, allCookies(cookies))
            .then(v=>{
                handlers.onCreate ? handlers.onCreate(v) : console.log("no onCreate provided")
            })
            .catch(e=>{
                setErr(JSON.stringify(e))
            })
    }
    return <NewEntryFormWrapper entryType={"project"}>
        <ErrorDisplay err={err}/>
        <div>
            <InputTextInlineTitle label={"Project Name"} readonly={false} value={name || ""} onChange={setName} />
        </div>
        <NewEntryNotes setNotes={setNotes}/>
        <button className={"greenButton buttonFullWidth"} onClick={createProject}>{"Create Project"}</button>
    </NewEntryFormWrapper>
}

export function ReadWriteAdminSelector({readonly, onUpdate, value}: {
    value: string,
    readonly: boolean,
    onUpdate?: (valOut: string) => void
}) {
    const strForVal = (str: string) => {
        return (str === "read") ? "can view" : (str === "admin" ? "is admin" : "can edit")
    }
    const valForStr = (str: string) => {
        return (str === "can edit") ? "write" : (str === "is admin" ? "admin" : "read")
    }
    if (readonly) {
        return <text>{strForVal(value)}</text>
    }
    return <SelectorFor options={["can view", "can edit", "is admin"]} initial={strForVal(value)}
                        updateParent={s => {
                            // if (s === "is admin") {
                            //     onUpdate && onUpdate("admin")
                            // }
                            // if (s === "can edit") {
                            //     onUpdate && onUpdate("write")
                            // }
                            // if (s === "can view") {
                            //     onUpdate && onUpdate("read")
                            // }
                            onUpdate && onUpdate(valForStr(s))
                            // TODO: what about the blank option???
                            return
                        }} disabled={false}/>
}

export function ReadWriteSelector({readonly, onUpdate, value}: {
    value: boolean,
    readonly: boolean,
    onUpdate?: (b: boolean) => void
}) {
    if (readonly) {
        return <text>{value ? "can edit" : "can view"}</text>
    }
    return <SelectorFor options={["can view", "can edit"]} initial={value ? "can edit" : "can view"}
                        updateParent={s => {
                            onUpdate && onUpdate(s === "can edit")
                        }} disabled={false}/>
}

export function ProjectPermsArea({perms, setPerms, readonly}: {
    perms?: Map<string, string>,
    setPerms?: (pp: Map<string, string>) => void,
    readonly: boolean,
}) {
    const [current, setCurrent] = useState(perms ? new Map<string, string>(perms) : new Map<string, string>())
    useEffect(() => {
        const temp = perms ? new Map<string, string>(perms) : new Map<string, string>()
        setCurrent(temp)
    }, [perms]);
    const depth = useContext(DepthContext)
    if (readonly && !current) {
        return null
    }
    // const updatePerms = (upd: Map<string,Data<string>>)=>{
    //     const toUpdateWith: Map<string, string> = new Map(
    //         [...upd.entries()]
    //             .filter(([k, v]) => !v.disabled)
    //             .map(([k, v]) => [k, v.data])
    //     );
    //     setPerms && setPerms(toUpdateWith)
    // }
    const existingUsersArea = ()=>{
        if (current === undefined || current.size === 0) {
            if (current === undefined){
                console.log("perms was undefined") // TODO: del
            } else {
                console.log("perms size was 0") // TODO: del
            }
            return null
        }
        return <>{/* TODO: make this into a grid or table?*/}
            {[...current.entries()].map(p => {
                return <>
                    <div key={p[0] + "name"}>{p[0]}</div>
                    <ReadWriteAdminSelector key={p[0] + "sel"} readonly={readonly} value={p[1]}
                                        onUpdate={(b) => {
                                            const updated = new Map<string, string>(current)
                                            setPerms && setPerms(updated.set(p[0], b))
                                        }}/>
                    <RemoveButton key={p[0] + "remv"} click={() => {
                        const updated = new Map<string, string>(current)
                        updated.delete(p[0])
                        setPerms && setPerms(updated)
                        // let updated = new Map<string, string>(perms)
                        // perms.entries().forEach(v => {
                        //     if (v[0] !== p[0]) {
                        //         updated.set(v[0], v[1])
                        //     }
                        // })
                        // setPerms && setPerms(updated)
                    }} txt={"Remove"}/>
                </>
        })}
        </>

    }
    return <DepthProvider>
        <div className={"subForm depth" + depth}>
            <div className={"centerH text-lg mb-1"}>{"Permissions"}</div>
            <div className={"projectPermsUsers"}>
                {existingUsersArea()}
            </div>

            {/* AREA TO ADD USER */}
            <div className={"inlineChildren"}>
                <div>{"Add user: "}</div>
                <UserSelector onSelect={(u) => {
                    const out = new Map<string, string>(current)
                    out.set(u._id, "read")
                    setPerms && setPerms(out)
                }} blacklist={(current !== undefined && current.size > 0) ? [...current.entries()].map(u => {
                return u[0]
            }) : []}/>
            </div>
        </div>
    </DepthProvider>
}

export async function GetMyProjects() { // TODO: use
    return await fetch(BaseExternalUrl + "/sessionUserProjects", { // TODO: ensure works like we want! We JUST want the user's perms on each project
        method: "GET",
        headers: clientPostRequestHeaders,
    }).then(HandleJsonResponse).then((projs) => {
        try {
            return projs as ProjectWithPerm[]
        } catch (err) {
            throw err
        }
    })
}

// TODO: used to return Promise<ProjectWithPerm[]>
export async function GetSessionUserProjects(complete?:boolean): Promise<string[]> {
    let params = ""
    if (complete !== undefined) {
        params = "?complete="+(complete?"true":"false")
    }
    return fetch(BaseExternalUrl + "/sessionUserProjects"+params, { // TODO: ensure works like we want! We JUST want the user's perms on each project
        method: "GET",
        headers: clientPostRequestHeaders,
    }).then(HandleJsonResponse).then((projs) => {
        try {
            return projs as string[] // TODO: ensure this return is correct?
        } catch (err) {
            throw err
        }
    })
}

// ProjectsSelector shows shows only projects that a user has associated with them
export function ProjectsSelector(inp: {
    onSelect: (projName: string) => void
    blacklist?: string[]
    complete?: boolean
}) {
    const [loading, setLoading] = useState(true)
    const [projects, setProjects] = useState<string[]>([])
    const [err, setErr] = useState<string | undefined>(undefined)
    useEffect(() => {
        GetSessionUserProjects(inp.complete).then(projNames=>{
            setProjects(projNames)
            setErr(undefined)
            setLoading(false)
        }).catch(e=>{
            HandleErr(e, setErr)
        })
    }, [])
    if (loading) {
        return <div>{"Loading projects selector"}</div>
    }
    const projectOptions = () => {
        if (projects == undefined || projects.length == 0) {
            return [""]
        }
        return ["", ...projects.filter(pToFilter => {
            return (inp.blacklist || []).indexOf(pToFilter) == -1
        })]
    }
    return <div>
        <ErrorDisplay err={err}/>
        <SelectorResetsOnSelectFor options={projectOptions()} updateParent={(pr) => {
            inp.onSelect(pr)
        }}/>
    </div>
}

export function ProjectListPageTable({data, onClick, withLink}: ListPageItems<ProjectData>) {
    let cols: ListTableColumn<ProjectData>[] = [
        NewColumn("Name", (v) => v._id, true),
        NewColumn("Completed", (v) => {
            return v.completed ? NumberToDateStr(v.completed) : ""
        }, true),
        NewColumn("Created", (v) => {
            return NumberToDateStr(v.creationDate)
        }, true),
        NewColumn("Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        }),
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: ProjectData)=>{
            return <EntryLinkWrapper props={{entry:v,openInNewTab:true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick} newClass={v=>{return new ProjectData(v)}}/>
}

export function ProjectSelectorTable({data, onClick}: ListPageItems<ProjectData>) {
    return <ProjectListPageTable data={data} onClick={onClick} withLink={true} />
}

// TODO: distinguish from ProjectsSelector
// ProjectSelector shows ALL projects, not just projects that a user has associated with them // TODO: is this ok?
export function ProjectSelector(
    {
        doSelect,
        allowCreate,
    }: {
        doSelect: (val: ProjectData | undefined) => void,
        allowCreate?: boolean
    }) {
    const table = (items: ProjectData[]):JSX.Element=>{
        return <ProjectSelectorTable data={items} onClick={doSelect}/>
    }
    const creationArea = ()=>{
        if(!allowCreate){
            return null
        }
        return <NewProjectForm handlers={{onCreate: doSelect,isTopLevel: false}}/>
    }

    return <ExistingRecentSelector entryType={"project"} entryTypes={"projects"} doSelect={doSelect} asserter={AssertProject} table={table}>
        {creationArea()}
    </ExistingRecentSelector>
}