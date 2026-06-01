'use client'

import React, {JSX, useContext, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import {
    AddCreatedTriColFunction,
    AllEntries,
    Data,
    OnViewCreatorQuadCol
} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import NutrientsArea, {
    IsValidNutrient,
    Nutrient,
    NutrientsEntriesGroupForNew
} from "@/app/components/formSubcomponents/nutrients";
import SugarsArea, {
    IsValidSugar,
    Sugar,
    SugarEntriesGroupForNew
} from "@/app/components/formSubcomponents/sugars";
import {
    createApiUrlFor,
    CreatedLinkFor, DisplayFormWrapper,
    DisplayInput, DoCreateRequest, DoUpdateRequest, ErrHandler, ExistingDualSelector, FlexedArea,
    FlexedSinglesGroup,
    HandleJsonResponse,
    HandleTxtResponse,
    ListPageItems, ListPageTable, ListTableColumn, NewColumn, NewEntryFormWrapper,
    NewEntryInput, NumberToDateStr,
    OptionalArrayOfType,
    OptionalKey,
    RequiredArrayOfType, updateApiUrlFor
} from "@/app/components/common";
import EntryLink, {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import AdditivesArea, {
    Additive,
    AdditiveEntriesGroupForNew,
    IsValidAdditive
} from "@/app/components/formSubcomponents/additives";
import {JarRecipeData} from "@/app/components/jarRecipeServer";
import {ErrorDisplay, NameArea, StandardArea} from "@/app/components/formSubcomponents/commonClient";
import {BaseExternalUrl} from "@/app/components/Constants";
import {Grain, GrainsSelector, IsValidGrain} from "@/app/components/formSubcomponents/grains";
import {AclDisplay, IsValidAcl, MarshalAcl, TogglableAreaWithDepth} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import TestAndValidate from "@/app/components/testing/untested";
import {AssertGrainBatch, NewGrainBatchForm} from "@/app/components/grainBatchClient";
import {GrainBatchData} from "@/app/components/grainBatchServer";
import {OnViewCreatorsTriColArea} from "@/app/components/formSubcomponents/ovc";
import {AssertJar} from "@/app/components/jarClient";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";


export function AssertJarRecipe(input: any): asserts input is JarRecipeData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    let requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['name', 'string'],
        ['standard', 'boolean'],
        ['lastUpdated', 'number'],
    ])
    for (let [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Grain Jar Recipe assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }

    // complex optional keys
    let complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['acl', IsValidAcl]
    ])
    for (let [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('Jar assertion failure: optional key ' + key + ' was not valid');
        }
    }

    // complex required array keys
    let complexRequiredArrayKeys = new Map<string, (v: any) => boolean>([
        ['grains', IsValidGrain], // TODO: ensure length > 0
    ])
    for (let [key, validator] of complexRequiredArrayKeys) {
        if (!RequiredArrayOfType(key, input, validator)) {
            throw new Error('Grain Jar Recipe assertion failure: required array key ' + key + ' was not valid');
        }
    }

    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['nutrients', IsValidNutrient],
        ['sugars', IsValidSugar],
        ['additives', IsValidAdditive],
        ['notes', IsValidNote],
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Grain Jar Recipe assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    return
}

function dataFormatFor<T>(init: T[] | undefined): Data<T>[] {
    return (init || []).map((v) => {
        return {data: v, disabled: false}
    })
}

export default function JarRecipeDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel
    }: DisplayInput) {
    try {
        AssertJarRecipe(data)
        const [initial, setInitial] = useState(data)
        // grain non-changeable (base grain)
        // name non-changeable
        const [isStandard, setIsStandard] = useState(initial.standard)
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
        const [acl, setAcl] = useState<ACL | undefined>(initial.acl)
        const [err, setErr] = useState<string | undefined>()
        const updateInitial = (updated: JarRecipeData) => {
            setInitial(updated)
            setIsStandard(updated.standard)
            setNotes(InitialNotesState(updated.notes))
            setAcl(updated.acl)
        }
        const cookies = useContext(CookiesContext)
        const submit = () => {
            const body: any = {
                standard: isStandard,
                notes: notes,
                acl: MarshalAcl(acl),
            }
            DoUpdateRequest("jarRecipe",initial._id, body, AssertJarRecipe, allCookies(cookies))
                .then(updateInitial)
                .catch(ErrHandler(setErr))
            // fetch(updateApiUrlFor("jarRecipe",data._id), {
            //     method: "POST",
            //     headers: clientPostRequestHeaders,
            //     body: JSON.stringify({
            //         standard: isStandard,
            //         notes: notes,
            //         acl: MarshalAcl(acl),
            //     })
            // })
            //     .then(HandleJsonResponse)
            //     .then((newEntry) => {
            //         AssertJarRecipe(newEntry)
            //         updateInitial(newEntry)
            //     })
            //     .catch(ErrHandler(setErr));
        }
        const jarGrainsArea = () => {
            return <div>
                {"Grains: "}
                {data.grains.map((g, i) => {
                    return <div key={g.grain}>{g.percentage + "% " + g.grain}</div>
                })}
            </div>
        }
        const ovcs: OnViewCreatorQuadCol[] = [
            {
                txt: "New Batch From Recipe",
                newCreationArea: (onCreate: AddCreatedTriColFunction) => {
                    return <NewGrainBatchForm recipe={initial} handlers={{
                        onCreate: (newItem: GrainBatchData) => {
                            return onCreate([{
                                typeText: "Grain Batch",
                                node: <CreatedLinkFor linkId={newItem._id} typ={"grainBatch"}/>
                            }], false)
                        },
                        isTopLevel: false,
                    }}/>
                },
            }
        ]
        return <DisplayFormWrapper entryType={"jarRecipe"}>
            <ErrorDisplay err={err} headerLevel={headerLevel}/>
            <TestAndValidate todos={["Put name at top????"]}>
                <ID id={data._id} txt={"Grain Jar Recipe"} entryType={"jarRecipe"}/>
            </TestAndValidate>
            <OnViewCreatorsTriColArea OnViewCreators={ovcs} readonly={readonly}/>
            <FlexedArea>
                <FlexedSinglesGroup>
                    <TestAndValidate todos={["allow to be changeable?", "sometimes deletes name when doing updates"]}>
                        <NameArea currentName={initial.name} readonly={readonly} headerLevel={headerLevel}/>
                    </TestAndValidate>

                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
                    <StandardArea isStandard={isStandard} readonly={readonly} setStandard={setIsStandard}
                                  headerLevel={headerLevel}/>
                </FlexedSinglesGroup>
            </FlexedArea>
            {jarGrainsArea()}

            <NutrientsArea initialValues={dataFormatFor(initial.nutrients)} headerLevel={headerLevel}
                           readonly={true}/>{/* TODO: make a viewOnlyArea?*/}
            <SugarsArea initialValues={dataFormatFor(initial.sugars)} headerLevel={headerLevel}
                        readonly={true}/>{/* TODO: make a viewOnlyArea?*/}
            <AdditivesArea initialValues={dataFormatFor(initial.additives)} headerLevel={headerLevel}
                           readonly={true}/>{/* TODO: make a viewOnlyArea?*/}
            <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
            <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                <AclDisplay ACL={acl} readonly={readonly} updateParent={setAcl}/>
            </TogglableAreaWithDepth>
            {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e) => {
                e.stopPropagation();
                submit()
            }}>{"Update"}</button>}

        </DisplayFormWrapper>
    } catch (err) {
        return <div>{"ERROR: Jar Recipe data format incorrect: " + err}</div>
    }
}

export function NewJarRecipeForm({handlers}: { handlers: NewEntryInput<JarRecipeData> }) {
    const [name, setName] = useState<string | undefined>()
    const [grains, setGrains] = useState<Grain[]>()
    const [isStandard, setIsStandard] = useState(false)
    const [nutrients, setNutrients] = useState<Nutrient[]>([])
    const [sugars, setSugars] = useState<Sugar[]>([])
    const [additives, setAdditives] = useState<Additive[]>([])
    const [notes, setNotes] = useState<Note[]>([])
    const [err, setErr] = useState<string | undefined>()
    const [templateSelectorOpen, setTemplateSelectorOpen] = useState<boolean>(false)
    const loadTemplate = (template: JarRecipeData) => {
        setGrains(template.grains)
        setNutrients(template.nutrients || [])
        setSugars(template.sugars || [])
        setAdditives(template.additives || [])
    }
    const errHandler = ErrHandler(setErr)
    const cookies = useContext(CookiesContext)
    const newJarRecipeSubmit = () => {
        if (!name) {
            setErr("Name must be set!")
            return
        }
        if (!grains || grains.length < 1) {
            setErr("Grains must be set!")
            return
        }
        let gs = grains || []
        let totalPct = 0
        for (let i = 0; i < gs.length; i++) {
            if (gs[i].percentage < 0 || gs[i].percentage > 100) {
                setErr("Invalid grain percentage")
            }
            totalPct += gs[i].percentage
        }
        if (totalPct != 100) {
            setErr("Grain percentages must equal 100")
            return
        }
        const body: any = {
            name: name,
            grain: grains,
            standard: isStandard,
            nutrients: nutrients,
            sugars: sugars,
            additives: additives,
            notes: notes,
        }
        DoCreateRequest("jarRecipe", body, AssertJarRecipe, allCookies(cookies))
            .then(handlers?.onCreate)
            .catch(errHandler)
        // fetch(createApiUrlFor("jarRecipe"), {
        //     method: 'Post',
        //     body: JSON.stringify({
        //         name: name,
        //         grain: grains,
        //         standard: isStandard,
        //         nutrients: nutrients,
        //         sugars: sugars,
        //         additives: additives,
        //         notes: notes,
        //     }),
        //     headers: clientPostRequestHeaders,
        // })
        //     .then(HandleJsonResponse)
        //     .then((newEntry) => {
        //         AssertJarRecipe(newEntry)
        //         handlers.onCreate && handlers.onCreate(newEntry)
        //     })
        //     .catch(ErrHandler(setErr));
    }
    const templateRecipeSelector = () => {
        if (templateSelectorOpen) {
            // TODO: closeable selector???
            return <JarRecipeSelector doSelect={(rec) => { // TODO: endpoint for getStandard?
                if (rec === undefined) {
                    return
                }
                loadTemplate(rec)
                setTemplateSelectorOpen(false)
            }} allowCreate={false}/>
        } else {
            return <button className={"basicButton"} onClick={() => {
                setTemplateSelectorOpen(true)
            }}>{"Select a template recipe (optional)"}</button>
        }
    }
    return <NewEntryFormWrapper entryType={"jarRecipe"}>
        <ErrorDisplay err={err}/>
        <TestAndValidate todos={["TEST THIS"]}>
            {templateRecipeSelector()}
        </TestAndValidate>
        <NameArea classNames={"inlineChildren"} readonly={false} setName={setName}/>

        <TestAndValidate todos={["ensure works like nutrientEntriesGroup"]}>
            <div>{"Grains"}</div>
            {/* TODO: grain batches???? */}
            <GrainsSelector current={grains || []} onChange={setGrains}/>
        </TestAndValidate>
        <StandardArea readonly={false} setStandard={setIsStandard}/>
        <div>{"Nutrients"}</div>
        <NutrientsEntriesGroupForNew currentEntries={nutrients} updateParent={setNutrients}/>
        <div>{"Sugars"}</div>
        <SugarEntriesGroupForNew currentEntries={sugars} updateParent={setSugars}/>
        <div>{"Additives"}</div>
        <AdditiveEntriesGroupForNew currentEntries={additives} updateParent={setAdditives}/>
        <NewEntryNotes setNotes={setNotes}/>
        <button className={"greenButton buttonFullWidth"} onClick={newJarRecipeSubmit}>{"Create Jar Recipe"}</button>
    </NewEntryFormWrapper>
}

// export function JarRecipeInline({
//                                     data,
//                                     expandByDefault,
//                                     onClick,
//                                     showMainPageButton,
//                                     idIsLink
//                                 }: InlineProps<JarRecipeData>) { // TODO: DO THIS ENTIRELY!
//     const [expanded, setExpanded] = useState(expandByDefault)
//     const b58id = data._id
//     return <InlineEntry onClick={onClick}>
//         <InlineSubArea props={{}}>
//             <ID id={b58id} txt={"Grain Jar Recipe"} entryType={"grainJar"} allowOpenMainPage={showMainPageButton}
//                 linkPage={idIsLink}/>
//             <NameArea currentName={data.name} readonly={true}/>
//             {/* TODO: ADD GRAINS */}
//             <StandardArea isStandard={data.standard} readonly={true}/>
//             <NutrientEntriesGroup preexisting={true} readonly={true} initialEntries={data.nutrients?.map((l) => {
//                 return {data: l, disabled: false}
//             })} updateParent={() => {
//             }}/>{/* TODO: NUTRIENTS (with more on expand)*/}
//             <SugarEntriesGroup preexisting={true} readonly={true} initialEntries={data.sugars?.map((l) => {
//                 return {data: l, disabled: false}
//             })} updateParent={() => {
//             }}/>{/* TODO: Liquids (with more on expand)*/}
//             <AdditiveEntriesGroup preexisting={true} readonly={true} initialEntries={data.additives?.map((l) => {
//                 return {data: l, disabled: false}
//             })} updateParent={() => {
//             }}/>{/* TODO: Additives (with more on expand) */}
//         </InlineSubArea>
//         <InlineExpansionArea props={{expanded: expanded}}>
//             <NotesAreaInline notes={data.notes}/>
//             <DateArea pre={"Last Updated: "} when={data.lastUpdated} readonly={true}/>
//         </InlineExpansionArea>
//         <InlineExpansionButton data-cy-id="InlineSubAreaButton" setExpanded={setExpanded}
//                                expanded={expanded}/>
//     </InlineEntry>
// }

export const JarRecipeArea = ({recipeId}: { recipeId?: string, headerLevel?: number }) => {
    let linkArea: JSX.Element | null = <div>{"unknown"}</div>
    if (recipeId !== undefined) {
        const b58id = recipeId
        linkArea = <EntryLink
            props={{displayedId: b58id, linkId: b58id, entryType: "jarRecipe"}}>{/* TODO: display name as well? */}
            <div>{b58id}</div>
            ]
        </EntryLink>
    }
    return <div>
        <div>{"Grain Recipe ID: "}</div>
        {linkArea}
    </div>
}

export function JarRecipeListPageTable({data, onClick, withLink}: ListPageItems<JarRecipeData>) {
    let cols: ListTableColumn<JarRecipeData>[] = [
        NewColumn("ID", (v) => v._id),
        NewColumn("Name", (v) => v.name), // TODO: shortname?
        NewColumn("Grains", (v) => {
            return <div>
                {v.grains.map((g, i) => {
                    return <div key={g.grain + i}>{g.grain}</div>
                })}
            </div>
        }),
        NewColumn("Nutrients", (v) => {
            return <div>
                {v.nutrients && v.nutrients.map((v, i) => {
                    return <div key={v.nutrient + i}>{v.nutrient}</div>
                })}
            </div>
        }),
        NewColumn("Sugars", (v) => {
            return <div>
                {v.sugars && v.sugars.map((v, i) => {
                    return <div key={v.type + i}>{v.type}</div>
                })}
            </div>
        }),
        NewColumn("Additives", (v) => {
            return <div>
                {v.additives && v.additives.map((v, i) => {
                    return <div key={v.additive + i}>{v.additive}</div>
                })}
            </div>
        }),
        NewColumn("Last Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        })
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: JarRecipeData) => {
            return <EntryLinkWrapper props={{linkId: encodeURI(v._id), entryType: "jarRecipe", openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable className={"text-xs"} cols={cols} data={data} onClick={onClick}/>
}

export function JarRecipeSelectorTable({data, onClick}: ListPageItems<JarRecipeData>) {
    let cols: ListTableColumn<JarRecipeData>[] = [
        NewColumn("Name", (v) => v.name), // TODO: shortname?
        NewColumn("ID", (v) => v._id),
        NewColumn("Last Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        }),
        NewColumn("Link", (v: JarRecipeData) => {
            return <EntryLinkWrapper props={{linkId: encodeURI(v._id), entryType: "jarRecipe", openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })
        // TODO: bonus area for notes???
    ]
    return <ListPageTable className={"text-xs"} cols={cols} data={data} onClick={onClick}/>
}

export function JarRecipeSelector(
    {
        doSelect,
        allowCreate,
        creatorInPage,
    }: {
        doSelect: (val: JarRecipeData | undefined) => void,
        allowCreate?: boolean,
        creatorInPage?: boolean
    }) {
    const table = (items: JarRecipeData[]): JSX.Element => {
        return <JarRecipeSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingDualSelector entryType={"jarRecipe"} entryTypes={"jarRecipes"} doSelect={doSelect}
                                 asserter={AssertJarRecipe}
                                 table={table}>
        {allowCreate && (creatorInPage ? <NewJarRecipeForm handlers={{onCreate: doSelect, isTopLevel: false}}/> :
            <TestAndValidate todos={["fix creator"]}><div>{"LINK TO CREATOR HERE FIXME"}</div></TestAndValidate>)}
    </ExistingDualSelector>
}